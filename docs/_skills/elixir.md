---
layout: skill
title: Elixir Patterns
name: elixir
language: elixir
description: Write idiomatic Elixir code following community best practices. Covers pattern matching, OTP, Phoenix, and functional programming patterns.
---

# Elixir Patterns

Guidelines for writing idiomatic Elixir and Phoenix applications.

## When to Activate

- Writing or editing Elixir code (`.ex`, `.exs` files)
- Working with Phoenix applications
- Reviewing Elixir code

## Pattern Matching

### Use pattern matching in function heads

```elixir
# GOOD - multiple function clauses
def handle_response({:ok, body}), do: process(body)
def handle_response({:error, reason}), do: log_error(reason)

# BAD - case inside function
def handle_response(response) do
  case response do
    {:ok, body} -> process(body)
    {:error, reason} -> log_error(reason)
  end
end
```

### Destructure in function arguments

```elixir
# GOOD
def create_user(%{"email" => email, "name" => name}) do
  %User{email: email, name: name}
end

# BAD
def create_user(params) do
  email = params["email"]
  name = params["name"]
  %User{email: email, name: name}
end
```

### Use guards for type checks

```elixir
def process(value) when is_binary(value), do: String.upcase(value)
def process(value) when is_integer(value), do: value * 2
def process(value) when is_list(value), do: Enum.sum(value)
```

## Pipe Operator

### Use pipes for data transformation

```elixir
# GOOD
result =
  data
  |> parse()
  |> validate()
  |> transform()
  |> save()

# BAD - nested calls
result = save(transform(validate(parse(data))))
```

### Start pipes with data, not function calls

```elixir
# GOOD
user
|> Map.get(:email)
|> String.downcase()

# BAD
Map.get(user, :email)
|> String.downcase()
```

## Error Handling

### Use tagged tuples

```elixir
# GOOD - consistent return types
def find_user(id) do
  case Repo.get(User, id) do
    nil -> {:error, :not_found}
    user -> {:ok, user}
  end
end

# Then pattern match on result
case find_user(id) do
  {:ok, user} -> render(conn, "show.html", user: user)
  {:error, :not_found} -> send_resp(conn, 404, "Not found")
end
```

### Use `with` for multiple operations

```elixir
# GOOD
def create_order(params) do
  with {:ok, user} <- find_user(params.user_id),
       {:ok, product} <- find_product(params.product_id),
       {:ok, order} <- build_order(user, product, params) do
    Repo.insert(order)
  end
end

# BAD - nested cases
def create_order(params) do
  case find_user(params.user_id) do
    {:ok, user} ->
      case find_product(params.product_id) do
        {:ok, product} ->
          # ... more nesting
      end
  end
end
```

## Phoenix

### Keep controllers thin

```elixir
defmodule MyAppWeb.OrderController do
  use MyAppWeb, :controller
  
  def create(conn, params) do
    case Orders.create(params, conn.assigns.current_user) do
      {:ok, order} ->
        conn
        |> put_flash(:info, "Order created")
        |> redirect(to: ~p"/orders/#{order}")
        
      {:error, changeset} ->
        render(conn, :new, changeset: changeset)
    end
  end
end
```

### Use contexts for business logic

```elixir
defmodule MyApp.Orders do
  alias MyApp.{Repo, Order}
  
  def create(attrs, user) do
    %Order{}
    |> Order.changeset(attrs)
    |> Ecto.Changeset.put_assoc(:user, user)
    |> Repo.insert()
  end
  
  def list_for_user(user) do
    Order
    |> where(user_id: ^user.id)
    |> order_by(desc: :inserted_at)
    |> Repo.all()
  end
end
```

### Use LiveView for interactive UIs

```elixir
defmodule MyAppWeb.OrderLive do
  use MyAppWeb, :live_view
  
  def mount(_params, _session, socket) do
    {:ok, assign(socket, orders: Orders.list_recent())}
  end
  
  def handle_event("delete", %{"id" => id}, socket) do
    Orders.delete(id)
    {:noreply, assign(socket, orders: Orders.list_recent())}
  end
end
```

## Ecto

### Use changesets for validation

```elixir
defmodule MyApp.User do
  use Ecto.Schema
  import Ecto.Changeset
  
  schema "users" do
    field :email, :string
    field :name, :string
    timestamps()
  end
  
  def changeset(user, attrs) do
    user
    |> cast(attrs, [:email, :name])
    |> validate_required([:email, :name])
    |> validate_format(:email, ~r/@/)
    |> unique_constraint(:email)
  end
end
```

### Use composable queries

```elixir
defmodule MyApp.UserQuery do
  import Ecto.Query
  
  def active(query \\ User), do: where(query, [u], u.active == true)
  def recent(query \\ User), do: order_by(query, [u], desc: u.inserted_at)
  def with_role(query \\ User, role), do: where(query, [u], u.role == ^role)
end

# Usage
User
|> UserQuery.active()
|> UserQuery.with_role(:admin)
|> UserQuery.recent()
|> Repo.all()
```

## OTP

### Use GenServer for stateful processes

```elixir
defmodule MyApp.Counter do
  use GenServer
  
  # Client API
  def start_link(initial \\ 0) do
    GenServer.start_link(__MODULE__, initial, name: __MODULE__)
  end
  
  def increment, do: GenServer.call(__MODULE__, :increment)
  def get, do: GenServer.call(__MODULE__, :get)
  
  # Server callbacks
  @impl true
  def init(initial), do: {:ok, initial}
  
  @impl true
  def handle_call(:increment, _from, count) do
    {:reply, count + 1, count + 1}
  end
  
  @impl true
  def handle_call(:get, _from, count) do
    {:reply, count, count}
  end
end
```

## Testing

### Use ExUnit with descriptive test names

```elixir
defmodule MyApp.OrdersTest do
  use MyApp.DataCase
  
  describe "create/2" do
    test "creates order with valid attrs" do
      user = insert(:user)
      attrs = %{product_id: 1, quantity: 2}
      
      assert {:ok, order} = Orders.create(attrs, user)
      assert order.user_id == user.id
    end
    
    test "returns error with invalid attrs" do
      user = insert(:user)
      attrs = %{}
      
      assert {:error, changeset} = Orders.create(attrs, user)
      assert errors_on(changeset) == %{product_id: ["can't be blank"]}
    end
  end
end
```
