
# Ruby Idioms

Guidelines for writing idiomatic, maintainable Ruby code.

## When to Activate

- Writing or editing Ruby code (`.rb` files)
- Creating new Ruby classes or modules
- Reviewing Ruby code

## Method Design

### Use keyword arguments for clarity

```ruby
# GOOD - clear intent
def create_user(name:, email:, admin: false)
  User.new(name: name, email: email, admin: admin)
end

# BAD - positional args are unclear
def create_user(name, email, admin = false)
  User.new(name: name, email: email, admin: admin)
end
```

### Use `fetch` for hash access with defaults

```ruby
# GOOD - explicit about default, raises on missing key
config.fetch(:timeout, 30)
config.fetch(:api_key) # raises KeyError if missing

# BAD - silent nil return hides bugs
config[:timeout] || 30
```

### Prefer guard clauses over nested conditionals

```ruby
# GOOD
def process(item)
  return unless item.valid?
  return if item.processed?
  
  # main logic here
end

# BAD
def process(item)
  if item.valid?
    unless item.processed?
      # main logic here
    end
  end
end
```

## Error Handling

### Use specific exception classes

```ruby
# GOOD
class ApiError < StandardError; end
class AuthenticationError < ApiError; end
class RateLimitError < ApiError; end

def call_api
  raise AuthenticationError, "Invalid API key"
end

# BAD
raise "Invalid API key"
```

### Wrap external calls

```ruby
# GOOD
def fetch_data
  response = HTTP.get(url)
  JSON.parse(response.body)
rescue HTTP::Error => e
  raise ApiError, "HTTP request failed: #{e.message}"
rescue JSON::ParserError => e
  raise ApiError, "Invalid JSON response: #{e.message}"
end
```

## Collections

### Use `map`/`select`/`reject` over `each` with mutation

```ruby
# GOOD
active_users = users.select(&:active?)
names = users.map(&:name)

# BAD
active_users = []
users.each { |u| active_users << u if u.active? }
```

### Use `each_with_object` for building hashes

```ruby
# GOOD
users.each_with_object({}) do |user, hash|
  hash[user.id] = user.name
end

# ALSO GOOD
users.to_h { |user| [user.id, user.name] }
```

## Modules and Classes

### Use modules for namespacing and mixins

```ruby
# Namespace
module MyApp
  module Services
    class PaymentProcessor
    end
  end
end

# Mixin
module Loggable
  def log(message)
    Rails.logger.info("[#{self.class.name}] #{message}")
  end
end
```

### Prefer composition over inheritance

```ruby
# GOOD - composition
class OrderProcessor
  def initialize(payment_gateway:, notifier:)
    @payment_gateway = payment_gateway
    @notifier = notifier
  end
end

# BAD - deep inheritance
class OrderProcessor < BaseProcessor < ApplicationService
end
```

## Testing Patterns

### Use `let` for lazy evaluation, `let!` when needed

```ruby
# Lazy - only evaluated when used
let(:user) { create(:user) }

# Eager - evaluated before each example
let!(:user) { create(:user) }
```

### Use `described_class` in specs

```ruby
# GOOD
RSpec.describe UserService do
  subject { described_class.new(user) }
end

# BAD
RSpec.describe UserService do
  subject { UserService.new(user) }
end
```

## Style

### Use `%w[]` and `%i[]` for arrays

```ruby
# GOOD
STATUSES = %w[pending active completed].freeze
ALLOWED_ROLES = %i[admin moderator user].freeze

# BAD
STATUSES = ['pending', 'active', 'completed'].freeze
```

### Use string interpolation over concatenation

```ruby
# GOOD
"Hello, #{user.name}!"

# BAD
'Hello, ' + user.name + '!'
```

### Freeze constants

```ruby
# GOOD
VALID_STATES = %w[draft published archived].freeze

# BAD - can be mutated
VALID_STATES = %w[draft published archived]
```
