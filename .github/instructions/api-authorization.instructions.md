---
description: "REST API Authorization Patterns and Security Guidelines"
applyTo: "services/api/routes/*.py,services/api/dependencies/permissions.py"
---

# REST API Authorization Patterns

## Core Principle

**All API endpoints must enforce proper authorization to prevent unauthorized access and information disclosure.**

## Authorization Architecture

The API uses a three-tier authorization model:

### 1. Authentication (services/api/dependencies/auth.py)

- `get_current_user()` - Validates session_token HTTPOnly cookie
- Returns CurrentUser with user model, access_token, and session_token
- **REQUIRED for ALL protected endpoints**

### 2. Permission Dependencies (services/api/dependencies/permissions.py)

- FastAPI Depends() injection for declarative authorization
- Wraps Discord permission checks and role verifications
- Provides reusable authorization logic across routes

### 3. Authorization Helpers (services/api/dependencies/permissions.py)

- Centralized functions for common authorization patterns
- Returns 404 for non-guild-members to prevent information disclosure
- Returns 403 only when user is guild member but lacks permissions

## Critical Security Pattern: 404 vs 403

**ALWAYS use 404 (Not Found) instead of 403 (Forbidden) when:**

- User is not a member of the guild containing the resource
- Resource exists but user shouldn't know it exists

**Why:** 403 reveals resource existence to unauthorized users. 404 prevents information leakage.

### Examples

```python
# ✅ CORRECT: Non-member receives 404
@router.get("/templates/{template_id}")
async def get_template(
    template_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    template = await verify_template_access(template_id, current_user, db)
    # Returns 404 if user not guild member
    return template

# ✅ CORRECT: Member without role receives 403
@router.post("/guilds/{guild_id}/templates")
async def create_template(
    guild_id: str,
    current_user: CurrentUser = Depends(require_bot_manager),
    # Raises 403 if user is guild member but lacks bot manager role
):
    pass

# ❌ WRONG: Don't use inline authorization checks
@router.get("/games/{game_id}")
async def get_game(game_id: str, current_user: CurrentUser = Depends(get_current_user)):
    game = await get_game_from_db(game_id)
    # DON'T DO THIS - use helper instead
    role_service = get_role_service()
    if not await role_service.has_permissions(...):
        raise HTTPException(status_code=403)
```

## Available Authorization Dependencies

### require_manage_guild

Requires Discord MANAGE_GUILD permission.

```python
from services.api.dependencies import permissions

@router.post("/guilds/{guild_id}/config")
async def create_config(
    guild_id: str,
    current_user: CurrentUser = Depends(permissions.require_manage_guild),
):
    pass
```

### require_manage_channels

Requires Discord MANAGE_CHANNELS permission.

```python
@router.post("/channels")
async def create_channel_config(
    current_user: CurrentUser = Depends(permissions.require_manage_channels),
):
    pass
```

### require_bot_manager

Requires bot manager role configured in guild settings.

```python
@router.post("/guilds/{guild_id}/templates")
async def create_template(
    guild_id: str,
    current_user: CurrentUser = Depends(permissions.require_bot_manager),
):
    pass
```

## Available Authorization Helpers

### verify_guild_membership

Verifies user is member of specified guild. Returns 404 if not member.

```python
from services.api.dependencies import permissions

@router.get("/guilds/{guild_id}")
async def get_guild(
    guild_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    user_guilds = await permissions.verify_guild_membership(guild_id, current_user, db)
    # Returns 404 if user not guild member
    # Returns list of user's guilds if member
```

### verify_template_access

Verifies user can access template (guild membership check). Returns 404 if not member.

```python
@router.get("/templates/{template_id}")
async def get_template(
    template_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    template = await permissions.verify_template_access(template_id, current_user, db)
    # Returns 404 if user not guild member
    # Returns template if authorized
```

### verify_game_access

Verifies user can access game (guild membership + player role checks). Returns 404 if not member, 403 if missing roles.

```python
@router.get("/games/{game_id}")
async def get_game(
    game_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    game = await permissions.verify_game_access(game_id, current_user, db)
    # Returns 404 if user not guild member
    # Returns 403 if user lacks required player roles
    # Returns game if authorized
```

### can_manage_game

Checks if user can modify game (host or bot manager).

```python
@router.put("/games/{game_id}")
async def update_game(
    game_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    game = await permissions.can_manage_game(game_id, current_user, db)
    # Returns 404 if user not guild member
    # Returns 403 if user not host or bot manager
    # Returns game if authorized
```

### can_export_game

Checks if user can export game (host, participant, or bot manager).

```python
@router.get("/export/game/{game_id}")
async def export_game(
    game_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    game = await permissions.can_export_game(game_id, current_user, db)
    # Returns 404 if user not guild member
    # Returns 403 if user not host, participant, or bot manager
    # Returns game if authorized
```

### get_guild_name

Centralized helper to fetch guild name from Discord API.

```python
from services.api.dependencies import permissions

guild_name = await permissions.get_guild_name(guild_id, current_user.access_token)
```

## MANDATORY Rules for New Endpoints

### ❌ DO NOT write inline authorization code

**NEVER do this:**

```python
# ❌ WRONG: Direct service calls in route handlers
role_service = get_role_service()
has_permission = await role_service.check_bot_manager_permission(...)
if not has_permission:
    raise HTTPException(status_code=403)

# ❌ WRONG: Direct Discord API calls in route handlers
user_guilds = await oauth2.get_user_guilds(current_user.access_token)
if guild_id not in [g["id"] for g in user_guilds]:
    raise HTTPException(status_code=403)
```

### ✅ DO use helper functions or dependencies

**ALWAYS do this:**

```python
# ✅ CORRECT: Use dependency for simple checks
current_user: CurrentUser = Depends(permissions.require_bot_manager)

# ✅ CORRECT: Use helper for resource-specific checks
template = await permissions.verify_template_access(template_id, current_user, db)
game = await permissions.verify_game_access(game_id, current_user, db)
```

## Guild Membership Verification

**CRITICAL: Always verify guild membership before exposing resources.**

### Why Guild Membership Matters

- Prevents users from accessing resources in guilds they don't belong to
- Prevents information disclosure about guild configuration
- Prevents enumeration of guilds/resources user shouldn't know about

### When to Verify Guild Membership

**Verify guild membership for:**

- Any endpoint returning guild-specific data
- Any endpoint returning templates (guild-scoped resources)
- Any endpoint returning games (guild-scoped resources)
- Any endpoint returning channel information
- Any endpoint modifying guild-scoped resources

**Example Pattern:**

```python
@router.get("/guilds/{guild_id}/data")
async def get_guild_data(
    guild_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    # ALWAYS verify guild membership first
    await permissions.verify_guild_membership(guild_id, current_user, db)
    
    # Now safe to return guild data
    return await fetch_guild_data(guild_id)
```

## Template and Game Role Restrictions

### Template Host Roles

Templates can restrict who can create games using `allowed_host_role_ids`.

**Enforcement:**

- Game creation: Use `require_game_host` dependency
- Verifies user has at least one of the allowed host roles
- Returns 403 if user lacks required roles

```python
@router.post("/games")
async def create_game(
    current_user: CurrentUser = Depends(permissions.require_game_host),
):
    pass
```

### Template Player Roles

Templates can restrict who can join games using `allowed_player_role_ids`.

**Enforcement:**

- Game list: Filter to only show games user can join
- Game detail: Verify user has required roles to view
- Game join: Verify user has required roles to join

```python
@router.get("/games")
async def list_games(
    guild_id: str,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(database.get_db),
):
    games = await get_all_games(guild_id)
    # Filter to only games user can access
    accessible_games = []
    for game in games:
        try:
            await permissions.verify_game_access(game.id, current_user, db)
            accessible_games.append(game)
        except HTTPException:
            continue  # User cannot access this game
    return accessible_games
```

## New Endpoint Authorization Checklist

When creating a new API endpoint, verify:

### Authentication

- [ ] Uses `Depends(get_current_user)` or stricter dependency
- [ ] Never exposes data to unauthenticated users (except public endpoints like /health, /auth/login)

### Guild Membership

- [ ] Verifies guild membership using `verify_guild_membership` helper
- [ ] Returns 404 (not 403) if user not guild member
- [ ] Never reveals existence of guilds user doesn't belong to

### Role-Based Authorization

- [ ] Uses appropriate dependency (`require_bot_manager`, `require_manage_guild`, `require_manage_channels`)
- [ ] OR uses helper function (`verify_template_access`, `verify_game_access`, `can_manage_game`, `can_export_game`)
- [ ] Returns 403 only when user is guild member but lacks permissions

### Template/Game Role Restrictions

- [ ] Game creation verifies template host roles
- [ ] Game list filters by template player roles
- [ ] Game detail verifies template player roles
- [ ] Game join verifies template player roles

### No Inline Authorization Code

- [ ] Zero direct RoleVerificationService calls in route handler
- [ ] Zero direct Discord API calls for guild membership in route handler
- [ ] Zero inline permission checks in route handler
- [ ] All authorization uses dependencies or helpers

### HTTP Status Codes

- [ ] 401 for unauthenticated requests
- [ ] 404 when resource doesn't exist OR user not guild member
- [ ] 403 when user is guild member but lacks permissions
- [ ] Never 403 for non-guild-members (use 404 to prevent information disclosure)

### Testing

- [ ] Unit tests for authorization helper usage
- [ ] Positive test case: authorized user succeeds
- [ ] Negative test case: non-member receives 404
- [ ] Negative test case: member without role receives 403
- [ ] Negative test case: unauthenticated user receives 401

## Security Rationale

### Information Disclosure Prevention

**Problem:** Returning 403 for resources in guilds user doesn't belong to reveals:

- Guild exists
- Resource exists (template, game, etc.)
- User could potentially gain access with different permissions

**Solution:** Return 404 for non-guild-members

- User cannot enumerate guilds
- User cannot discover resources in unknown guilds
- Only members see 403 for permission issues

### Centralized Authorization

**Problem:** Inline authorization code leads to:

- Inconsistent authorization logic across endpoints
- Authorization bugs from copy-paste errors
- Difficult security audits
- Hard to maintain and update

**Solution:** Use helper functions and dependencies

- Single source of truth for each authorization pattern
- Consistent behavior across all endpoints
- Easy to audit and test
- Changes propagate automatically

## Summary

**Three Rules for Secure APIs:**

1. **ALWAYS verify guild membership first** - Return 404 for non-members
2. **NEVER write inline authorization code** - Use helpers and dependencies
3. **ONLY return 403 to guild members** - Use 404 to prevent information disclosure

Following these patterns ensures:

- No unauthorized access
- No information disclosure
- Consistent authorization behavior
- Maintainable security posture
- Easy security audits
