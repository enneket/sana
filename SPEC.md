# Sana — Self-hosted Notes App

## 1. Concept & Vision

Minimalist self-hosted notes app with folder-first organization. Clean, fast, no cloud dependency. Your notes live on your server, accessed via browser. Feels like a local app, works like a web app.

## 2. Design Language

- **Aesthetic**: Minimal, calm, distraction-free. Inspired by iA Writer and Obsidian's clean mode.
- **Colors**: Monochrome with a single accent. `#1a1a2e` bg, `#e8e8e8` text, `#4a9eff` accent.
- **Typography**: `JetBrains Mono` for code/metadata, `Inter` for UI, system serif for note content.
- **Spacing**: 8px base unit. Generous whitespace.
- **Motion**: Minimal — 150ms transitions on hover, no decorative animation.
- **Icons**: Lucide icons (outline style).

## 3. Tech Stack

- **Backend**: Go, single binary, SQLite for storage, JWT auth
- **Frontend**: Vue 3 (Vite), plain CSS, no UI framework
- **Storage**: Filesystem for markdown files, SQLite for metadata/auth
- **Auth**: JWT tokens, httpOnly cookies

## 4. Features

### 4.1 Folder Management
- Create/rename/delete folders
- Nested folders (max 3 levels deep)
- Folder tree sidebar

### 4.2 Note Management
- Create/edit/delete markdown notes
- Notes belong to a folder
- Auto-save (debounced 1s)
- Basic markdown preview (not full editor)

### 4.3 Auth
- Register/login with username + password
- JWT-based sessions
- Logout

## 5. Layout

```
┌─────────────────────────────────────────────────┐
│  Header: logo + user menu                        │
├──────────────┬──────────────────────────────────┤
│  Sidebar     │  Main Content                    │
│  - Folders   │  - Note list (folder view)      │
│  - New folder│  - Note editor                   │
│              │                                  │
└──────────────┴──────────────────────────────────┘
```

## 6. API Endpoints

### Auth
- `POST /api/auth/register` — `{username, password}` → `{token}`
- `POST /api/auth/login` — `{username, password}` → `{token}`
- `POST /api/auth/logout`
- `GET /api/auth/me` → `{user}`

### Folders
- `GET /api/folders` → `[{id, name, parent_id}]`
- `POST /api/folders` — `{name, parent_id?}`
- `PUT /api/folders/:id` — `{name}`
- `DELETE /api/folders/:id`

### Notes
- `GET /api/notes?folder_id=` → `[{id, title, folder_id, updated_at}]`
- `GET /api/notes/:id` → `{id, title, content, folder_id}`
- `POST /api/notes` — `{title, content, folder_id}`
- `PUT /api/notes/:id` — `{title?, content?}`
- `DELETE /api/notes/:id`

## 7. Data Model

### Users (SQLite)
```sql
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Folders (SQLite)
```sql
CREATE TABLE folders (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  parent_id TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (parent_id) REFERENCES folders(id)
);
```

### Notes metadata (SQLite)
```sql
CREATE TABLE notes (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  folder_id TEXT NOT NULL,
  title TEXT NOT NULL,
  filename TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (folder_id) REFERENCES folders(id)
);
```

### Note content
Markdown files stored at `notes/{user_id}/{folder_id}/{note_id}.md`

## 8. Implementation Order

1. Go backend: project setup, SQLite init, JWT auth
2. Go backend: folder CRUD
3. Go backend: note CRUD + file storage
4. Vue 3 frontend: project setup, router, auth flow
5. Vue 3 frontend: folder sidebar
6. Vue 3 frontend: note list + editor
7. Docker setup
