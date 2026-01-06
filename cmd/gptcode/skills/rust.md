---
layout: skill
title: Rust Patterns
name: rust
language: rust
description: Write idiomatic Rust code with proper ownership, error handling, and patterns.
---

# Rust Patterns

Guidelines for writing safe, performant Rust code.

## When to Activate

- Writing or editing Rust code (`.rs` files)
- Creating Rust crates or modules
- Reviewing Rust code

## Ownership and Borrowing

### Prefer borrowing over ownership transfer

```rust
// GOOD - borrows the string
fn print_length(s: &str) {
    println!("Length: {}", s.len());
}

// BAD - takes ownership unnecessarily
fn print_length(s: String) {
    println!("Length: {}", s.len());
}
```

### Use &str for string parameters, String for owned data

```rust
// GOOD
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

// Works with both &str and String
greet("Alice");
greet(&my_string);
```

### Clone only when necessary

```rust
// GOOD - pass reference when possible
fn process(data: &[u8]) { ... }

// Clone when you need owned copy
let copy = data.clone();

// BAD - unnecessary clone
fn process(data: Vec<u8>) { ... }  // Takes ownership
process(data.clone());  // Forced to clone
```

## Error Handling

### Use Result for recoverable errors

```rust
use std::fs::File;
use std::io::{self, Read};

fn read_file(path: &str) -> Result<String, io::Error> {
    let mut file = File::open(path)?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;
    Ok(contents)
}
```

### Use thiserror for custom errors

```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("User {0} not found")]
    NotFound(u64),
    
    #[error("Invalid input: {0}")]
    Validation(String),
    
    #[error("Database error")]
    Database(#[from] sqlx::Error),
    
    #[error("IO error")]
    Io(#[from] std::io::Error),
}
```

### Use anyhow for applications, thiserror for libraries

```rust
// Application code - use anyhow
use anyhow::{Context, Result};

fn main() -> Result<()> {
    let config = load_config()
        .context("Failed to load configuration")?;
    run_app(config)?;
    Ok(())
}

// Library code - use thiserror
#[derive(thiserror::Error, Debug)]
pub enum LibError { ... }
```

## Option Handling

### Use combinators instead of match

```rust
// GOOD - combinators
let name = user.map(|u| u.name).unwrap_or("Anonymous".to_string());

let result = items
    .first()
    .filter(|item| item.active)
    .map(|item| item.value);

// OK for complex logic
match user {
    Some(u) if u.verified => process(u),
    Some(u) => verify_and_process(u),
    None => handle_anonymous(),
}
```

### Use if let for single-arm matches

```rust
// GOOD
if let Some(user) = find_user(id) {
    process(user);
}

if let Err(e) = operation() {
    log::error!("Failed: {}", e);
}

// BAD
match find_user(id) {
    Some(user) => process(user),
    None => {},
}
```

## Structs and Enums

### Use builder pattern for complex construction

```rust
#[derive(Default)]
pub struct Request {
    url: String,
    method: Method,
    headers: HashMap<String, String>,
    timeout: Duration,
}

impl Request {
    pub fn new(url: impl Into<String>) -> Self {
        Self {
            url: url.into(),
            ..Default::default()
        }
    }
    
    pub fn method(mut self, method: Method) -> Self {
        self.method = method;
        self
    }
    
    pub fn header(mut self, key: &str, value: &str) -> Self {
        self.headers.insert(key.to_string(), value.to_string());
        self
    }
}

// Usage
let req = Request::new("https://api.example.com")
    .method(Method::POST)
    .header("Content-Type", "application/json");
```

### Use newtype pattern for type safety

```rust
// GOOD - distinct types
struct UserId(u64);
struct OrderId(u64);

fn find_user(id: UserId) -> Option<User> { ... }
fn find_order(id: OrderId) -> Option<Order> { ... }

// BAD - can mix up IDs
fn find_user(id: u64) -> Option<User> { ... }
fn find_order(id: u64) -> Option<Order> { ... }
```

## Iterators

### Use iterator adapters

```rust
// GOOD
let active_names: Vec<_> = users
    .iter()
    .filter(|u| u.active)
    .map(|u| &u.name)
    .collect();

let total: u64 = orders.iter().map(|o| o.amount).sum();

// BAD
let mut active_names = Vec::new();
for user in &users {
    if user.active {
        active_names.push(&user.name);
    }
}
```

### Use collect into specific types

```rust
// Collect into HashMap
let map: HashMap<_, _> = items
    .iter()
    .map(|i| (i.id, i.name.clone()))
    .collect();

// Collect Results
let results: Result<Vec<_>, _> = items
    .iter()
    .map(|i| parse(i))
    .collect();
```

## Traits

### Implement standard traits

```rust
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct User {
    pub id: u64,
    pub name: String,
}

// Implement Display for user-facing output
impl std::fmt::Display for User {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}#{}", self.name, self.id)
    }
}
```

### Use impl Trait for return types

```rust
// GOOD - hides implementation detail
fn active_users(users: &[User]) -> impl Iterator<Item = &User> {
    users.iter().filter(|u| u.active)
}

// Also for accepting closures
fn map_all<F>(items: &[Item], f: F) -> Vec<String>
where
    F: Fn(&Item) -> String,
{
    items.iter().map(f).collect()
}
```

## Async

### Use async/await with tokio or async-std

```rust
use tokio;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let response = reqwest::get("https://api.example.com/data").await?;
    let data: ApiResponse = response.json().await?;
    println!("{:?}", data);
    Ok(())
}
```

### Use join for concurrent operations

```rust
use tokio::join;

async fn fetch_all() -> Result<(Users, Orders), Error> {
    let (users, orders) = join!(
        fetch_users(),
        fetch_orders()
    );
    Ok((users?, orders?))
}
```
