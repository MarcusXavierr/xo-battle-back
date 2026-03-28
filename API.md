# XO-Battle API Documentation

## Routes

### GET `/` — Health Check

Returns a plain text response to verify the server is running.

**Response:** `200 OK` — `"The system is alive"`

---

### POST `/room/{id}` — Create Room

Creates a new game room.

| Parameter | In  | Description       |
| --------- | --- | ----------------- |
| `id`      | URL | Room name/identifier |

**Responses:**

- `200` — `"Created room: {id}"`
- `200` — `"The room '{id}' already exists dipshit"` (if already exists)

---

### GET `/room/{id}/join` — Join Room (WebSocket)

Upgrades to a WebSocket connection and joins the player to an existing room.

| Parameter     | In    | Required | Description                                                                 |
| ------------- | ----- | -------- | --------------------------------------------------------------------------- |
| `id`          | URL   | yes      | Room name to join                                                           |
| `name`        | query | yes      | Player display name                                                         |
| `player_type` | query | no       | `"x"` or `"o"` (case-insensitive). First player only — ignored for second. |

**Type assignment rules:**

- **First player:** if `player_type` is provided and valid, that type is assigned. If omitted, defaults to `"x"`. Invalid value returns `400`.
- **Second player:** `player_type` is ignored. The server assigns the opposite of the first player's type.

Clients can infer their own type from the `player_joined` message: it is the opposite of the opponent's `player_type`.

**Error Responses:**

- `400` — `"Missing name"`
- `400` — `"invalid player type"` (first player sent an unrecognized value)

---

## WebSocket Protocol

**Connection URL:** `ws://<host>/room/{id}/join?name={name}` (first player may optionally add `&player_type={x|o}`)

All messages are JSON text frames.

---

### Messages: Server → Client

#### `player_joined`

Sent to both players when the second player connects.

```json
{
  "type": "player_joined",
  "name": "opponent_name",
  "player_type": "x",
  "order": 1
}
```

| Field         | Type     | Description                              |
| ------------- | -------- | ---------------------------------------- |
| `type`        | string   | `"player_joined"`                        |
| `name`        | string   | Opponent's display name                  |
| `player_type` | string   | Opponent's symbol (`"x"` or `"o"`)      |
| `order`       | int      | Join order — `1` for first, `2` for second |

#### `player_disconnected`

Sent when the opponent closes their connection. Your connection is also closed after this message.

```json
{
  "type": "player_disconnected"
}
```

#### `move`

Relayed from the opponent when they make a move.

```json
{
  "type": "move",
  "cell": 4
}
```

| Field  | Type   | Description                          |
| ------ | ------ | ------------------------------------ |
| `type` | string | `"move"`                             |
| `cell` | int    | Board cell index (`0`–`8`)           |

---

### Messages: Client → Server

#### `move`

Send a move to your opponent. The server relays it without validation.

```json
{
  "type": "move",
  "cell": 4
}
```

| Field  | Type   | Description                          |
| ------ | ------ | ------------------------------------ |
| `type` | string | `"move"`                             |
| `cell` | int    | Board cell index (`0`–`8`)           |

> **Note:** The server does not validate move legality — that is the client's responsibility.

---

## Message Flow

```
Player A                     Server                    Player B
   |                           |                          |
   |--- POST /room/game ------>|                          |
   |<-- 200 "Created room" ----|                          |
   |                           |                          |
   |--- WS /room/game/join --->|                          |
   |    ?name=A&player_type=x  |                          |
   |    (or just ?name=A)       |                          |
   |<-- connection established  |                          |
   |                           |                          |
   |                           |<-- WS /room/game/join ---|
   |                           |    ?name=B               |
   |                           |    (player_type ignored)  |
   |                           |--- connection established>|
   |                           |                          |
   |<-- player_joined ---------|--- player_joined ------->|
   |    {name:B, type:o, #2}   |    {name:A, type:x, #1} |
   |                           |                          |
   |--- move {cell:0} -------->|--- move {cell:0} ------>|
   |                           |                          |
   |<-- move {cell:4} ---------|<-- move {cell:4} -------|
   |                           |                          |
   |--- disconnect ----------->|                          |
   |                           |--- player_disconnected ->|
   |                           |--- close connection ---->|
```
