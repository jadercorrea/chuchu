---
layout: skill
title: TypeScript Patterns
name: typescript
language: typescript
description: Write idiomatic TypeScript code with proper type safety and modern patterns.
---

# TypeScript Patterns

Guidelines for writing type-safe, maintainable TypeScript code.

## When to Activate

- Writing or editing TypeScript code (`.ts`, `.tsx` files)
- Creating new modules or components
- Reviewing TypeScript code

## Type Safety

### Use strict mode

```typescript
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true
  }
}
```

### Prefer interfaces for objects, types for unions

```typescript
// GOOD - interface for objects
interface User {
  id: number;
  name: string;
  email: string;
}

// GOOD - type for unions/intersections
type Status = 'pending' | 'active' | 'completed';
type UserWithRole = User & { role: Role };

// BAD - type for simple objects
type User = {
  id: number;
  name: string;
};
```

### Use unknown instead of any

```typescript
// GOOD - type-safe
function parseJSON(json: string): unknown {
  return JSON.parse(json);
}

const data = parseJSON(input);
if (isUser(data)) {
  console.log(data.name); // Type-safe access
}

// BAD - loses type safety
function parseJSON(json: string): any {
  return JSON.parse(json);
}
```

### Use const assertions for literals

```typescript
// GOOD - immutable and precise types
const CONFIG = {
  api: 'https://api.example.com',
  timeout: 5000,
} as const;

const ROLES = ['admin', 'user', 'guest'] as const;
type Role = typeof ROLES[number]; // 'admin' | 'user' | 'guest'

// BAD - mutable and wide types
const CONFIG = {
  api: 'https://api.example.com', // string, not literal
  timeout: 5000, // number, not 5000
};
```

## Functions

### Use explicit return types for public APIs

```typescript
// GOOD - explicit return type
function createUser(data: CreateUserDTO): Promise<User> {
  return db.users.create(data);
}

// OK for internal/simple functions - inferred
const double = (n: number) => n * 2;
```

### Use function overloads for complex signatures

```typescript
// GOOD - overloaded
function find(id: number): User | undefined;
function find(email: string): User | undefined;
function find(idOrEmail: number | string): User | undefined {
  if (typeof idOrEmail === 'number') {
    return users.find(u => u.id === idOrEmail);
  }
  return users.find(u => u.email === idOrEmail);
}
```

### Use generics for reusable code

```typescript
// GOOD
function first<T>(items: T[]): T | undefined {
  return items[0];
}

async function fetchData<T>(url: string): Promise<T> {
  const response = await fetch(url);
  return response.json();
}

// Usage with type inference
const user = await fetchData<User>('/api/user');
```

## Error Handling

### Use Result types for expected failures

```typescript
type Result<T, E = Error> = 
  | { success: true; data: T }
  | { success: false; error: E };

async function findUser(id: number): Promise<Result<User, 'not_found' | 'db_error'>> {
  try {
    const user = await db.users.find(id);
    if (!user) {
      return { success: false, error: 'not_found' };
    }
    return { success: true, data: user };
  } catch {
    return { success: false, error: 'db_error' };
  }
}

// Usage
const result = await findUser(123);
if (result.success) {
  console.log(result.data.name);
} else {
  console.error(result.error);
}
```

### Use custom error classes

```typescript
class NotFoundError extends Error {
  constructor(public resource: string, public id: string | number) {
    super(`${resource} with id ${id} not found`);
    this.name = 'NotFoundError';
  }
}

class ValidationError extends Error {
  constructor(public field: string, public message: string) {
    super(`${field}: ${message}`);
    this.name = 'ValidationError';
  }
}
```

## Async/Await

### Always handle promise rejections

```typescript
// GOOD
try {
  const data = await fetchData();
} catch (error) {
  if (error instanceof NetworkError) {
    showRetryButton();
  } else {
    throw error;
  }
}

// BAD - unhandled rejection
const data = await fetchData(); // Will crash if rejected
```

### Use Promise.all for parallel operations

```typescript
// GOOD - parallel
const [users, orders] = await Promise.all([
  fetchUsers(),
  fetchOrders(),
]);

// BAD - sequential (slower)
const users = await fetchUsers();
const orders = await fetchOrders();
```

## React Specific (TSX)

### Type component props explicitly

```typescript
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary';
  disabled?: boolean;
}

function Button({ label, onClick, variant = 'primary', disabled }: ButtonProps) {
  return (
    <button 
      className={`btn-${variant}`}
      onClick={onClick}
      disabled={disabled}
    >
      {label}
    </button>
  );
}
```

### Use discriminated unions for state

```typescript
type State =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: User[] }
  | { status: 'error'; error: string };

function UserList() {
  const [state, setState] = useState<State>({ status: 'idle' });
  
  if (state.status === 'loading') return <Spinner />;
  if (state.status === 'error') return <Error message={state.error} />;
  if (state.status === 'success') return <List users={state.data} />;
  return <Button onClick={load}>Load</Button>;
}
```

## Testing

### Use proper typing in tests

```typescript
import { describe, it, expect, vi } from 'vitest';

describe('UserService', () => {
  it('creates user with valid data', async () => {
    const mockDb = {
      users: {
        create: vi.fn().mockResolvedValue({ id: 1, name: 'Test' }),
      },
    };
    
    const service = new UserService(mockDb as unknown as Database);
    const user = await service.create({ name: 'Test' });
    
    expect(user.id).toBe(1);
    expect(mockDb.users.create).toHaveBeenCalledWith({ name: 'Test' });
  });
});
```
