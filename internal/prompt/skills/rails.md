
# Rails Patterns

Guidelines for writing Rails applications following community best practices.

## When to Activate

- Working with Rails applications
- Creating controllers, models, or services
- Writing RSpec tests for Rails

## Model Patterns

### Use scopes for reusable queries

```ruby
class User < ApplicationRecord
  scope :active, -> { where(active: true) }
  scope :recent, -> { order(created_at: :desc) }
  scope :with_role, ->(role) { where(role: role) }
end

# Usage
User.active.recent.with_role(:admin)
```

### Validate at the right level

```ruby
class Order < ApplicationRecord
  # Database-enforced constraints
  validates :email, presence: true
  validates :amount, numericality: { greater_than: 0 }
  
  # Business logic validation
  validate :inventory_available, on: :create
  
  private
  
  def inventory_available
    return if line_items.all?(&:in_stock?)
    errors.add(:base, "Some items are out of stock")
  end
end
```

### Use callbacks sparingly

```ruby
# GOOD - simple, side-effect free
before_validation :normalize_email

# BAD - external side effects in callbacks
after_create :send_welcome_email  # Move to service/job
after_save :sync_to_external_api  # Move to service/job
```

## Controller Patterns

### Keep controllers thin

```ruby
class OrdersController < ApplicationController
  def create
    result = CreateOrder.call(order_params, current_user)
    
    if result.success?
      redirect_to result.order, notice: "Order created"
    else
      @order = result.order
      render :new, status: :unprocessable_entity
    end
  end
  
  private
  
  def order_params
    params.require(:order).permit(:product_id, :quantity)
  end
end
```

### Use strong parameters correctly

```ruby
def user_params
  params.require(:user).permit(
    :name, 
    :email,
    address_attributes: [:street, :city, :zip]
  )
end
```

## Service Objects

### Use a consistent interface

```ruby
class CreateOrder
  def self.call(...)
    new(...).call
  end
  
  def initialize(params, user)
    @params = params
    @user = user
  end
  
  def call
    order = Order.new(@params.merge(user: @user))
    
    if order.save
      NotifyOrderJob.perform_later(order)
      Result.success(order: order)
    else
      Result.failure(order: order, errors: order.errors)
    end
  end
end
```

### Keep services focused

```ruby
# GOOD - single responsibility
class ProcessPayment
  def call(order)
    # Only payment processing logic
  end
end

class SendOrderConfirmation
  def call(order)
    # Only notification logic
  end
end

# BAD - doing too much
class CreateOrder
  def call
    validate_inventory
    calculate_totals
    process_payment
    send_confirmation
    update_analytics
    sync_to_warehouse
  end
end
```

## Testing with RSpec

### Use factories, not fixtures

```ruby
# spec/factories/users.rb
FactoryBot.define do
  factory :user do
    name { Faker::Name.name }
    email { Faker::Internet.email }
    
    trait :admin do
      role { :admin }
    end
  end
end

# Usage
create(:user, :admin)
```

### Test behavior, not implementation

```ruby
# GOOD - tests behavior
it "creates an order for the user" do
  expect { described_class.call(params, user) }
    .to change(user.orders, :count).by(1)
end

# BAD - tests implementation
it "calls Order.create with params" do
  expect(Order).to receive(:create).with(params)
  described_class.call(params, user)
end
```

### Use request specs for APIs

```ruby
RSpec.describe "Orders API", type: :request do
  describe "POST /api/orders" do
    it "creates an order" do
      post "/api/orders", params: { order: valid_params }
      
      expect(response).to have_http_status(:created)
      expect(json_response[:id]).to be_present
    end
  end
end
```

## Database

### Add indexes for foreign keys and queried columns

```ruby
class CreateOrders < ActiveRecord::Migration[7.0]
  def change
    create_table :orders do |t|
      t.references :user, null: false, foreign_key: true
      t.string :status, null: false, default: "pending"
      t.timestamps
    end
    
    add_index :orders, :status
    add_index :orders, [:user_id, :status]
  end
end
```

### Avoid N+1 queries

```ruby
# GOOD - eager loading
def index
  @orders = Order.includes(:user, :line_items).recent
end

# BAD - N+1
def index
  @orders = Order.recent
  # Each order.user triggers a query
end
```

## Background Jobs

### Use jobs for slow operations

```ruby
class SendWelcomeEmailJob < ApplicationJob
  queue_as :default
  
  def perform(user_id)
    user = User.find(user_id)
    UserMailer.welcome(user).deliver_now
  end
end

# Usage - pass IDs, not objects
SendWelcomeEmailJob.perform_later(user.id)
```

### Make jobs idempotent

```ruby
class ProcessPaymentJob < ApplicationJob
  def perform(order_id)
    order = Order.find(order_id)
    
    # Skip if already processed
    return if order.paid?
    
    PaymentGateway.charge(order)
    order.update!(status: :paid)
  end
end
```
