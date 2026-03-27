mod task;
mod storage;
mod user;
mod session;

use std::env;
use std::process;
use chrono::Utc;

use task::{Task, parse_state};
use storage::Storage;
use user::{User, hash_sha256};
use session::{save_session, load_session, clear_session};

const DATA_FILE: &str = "tasks.json";

/// Punto de entrada del programa.
/// Los comandos de autenticación (register, login, logout) no requieren sesión activa.
/// Los comandos de tareas (add, list, update, delete, status) requieren una sesión válida.
fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        print_usage();
        process::exit(1);
    }

    let storage = Storage::new(DATA_FILE.to_string());
    let command = &args[1];
    let cmd_args = &args[2..];

    // Comandos de autenticación (no requieren sesión)
    match command.as_str() {
        "register" => { handle_register(&storage, cmd_args); return; }
        "login" => { handle_login(&storage, cmd_args); return; }
        "logout" => { handle_logout(); return; }
        _ => {}
    }

    // Comandos de tareas (requieren sesión activa)
    let store = storage.load().unwrap_or_else(|e| {
        eprintln!("Error loading data: {}", e);
        process::exit(1);
    });

    let user = load_session(&store.users).unwrap_or_else(|e| {
        eprintln!("Error: {}", e);
        process::exit(1);
    });

    match command.as_str() {
        "add" => handle_add(&storage, cmd_args, &user),
        "list" => handle_list(&storage, cmd_args, &user),
        "update" => handle_update(&storage, cmd_args, &user),
        "delete" => handle_delete(&storage, cmd_args, &user),
        "status" => handle_status(&storage, cmd_args, &user),
        _ => {
            eprintln!("Unknown command: {}", command);
            print_usage();
            process::exit(1);
        }
    }
}

fn print_usage() {
    println!("Task Tracker CLI");
    println!();
    println!("Auth:");
    println!("  task-tracker register <username> <password>");
    println!("  task-tracker login <username> <password>");
    println!("  task-tracker logout");
    println!();
    println!("Tasks (requires login):");
    println!("  task-tracker add <name> <description>");
    println!("  task-tracker list [state]");
    println!("  task-tracker update <id> <name> <description>");
    println!("  task-tracker delete <id>");
    println!("  task-tracker status <id> <state>");
    println!();
    println!("States: PENDING, IN_PROGRESS, DONE");
}

// ============================================================
// HANDLERS DE AUTENTICACIÓN
// ============================================================

fn handle_register(s: &Storage, args: &[String]) {
    if args.len() < 2 {
        eprintln!("Error: 'register' requires <username> and <password>");
        eprintln!("Usage: task-tracker register <username> <password>");
        process::exit(1);
    }

    let mut store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading data: {}", e);
        process::exit(1);
    });

    if s.find_user_by_username(&store, &args[0]).is_ok() {
        eprintln!("Error: username {:?} is already taken", &args[0]);
        process::exit(1);
    }

    let user = User::new(store.next_user_id, args[0].clone(), args[1].clone());
    println!("User {:?} registered successfully (ID: {})", user.username, user.id);

    store.users.push(user);
    store.next_user_id += 1;

    s.save(&store).unwrap_or_else(|e| {
        eprintln!("Error saving data: {}", e);
        process::exit(1);
    });
}

fn handle_login(s: &Storage, args: &[String]) {
    if args.len() < 2 {
        eprintln!("Error: 'login' requires <username> and <password>");
        eprintln!("Usage: task-tracker login <username> <password>");
        process::exit(1);
    }

    let store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading data: {}", e);
        process::exit(1);
    });

    let user = match s.find_user_by_username(&store, &args[0]) {
        Ok(u) => u,
        Err(_) => {
            eprintln!("Error: invalid username or password");
            process::exit(1);
        }
    };

    if user.password != hash_sha256(&args[1]) {
        eprintln!("Error: invalid username or password");
        process::exit(1);
    }

    save_session(user.id).unwrap_or_else(|e| {
        eprintln!("Error creating session: {}", e);
        process::exit(1);
    });

    println!("Welcome back, {}!", user.username);
}

fn handle_logout() {
    clear_session().unwrap_or_else(|e| {
        eprintln!("Error: {}", e);
        process::exit(1);
    });

    println!("Logged out successfully.");
}

// ============================================================
// HANDLERS DE TAREAS
// ============================================================

fn handle_add(s: &Storage, args: &[String], user: &User) {
    if args.len() < 2 {
        eprintln!("Error: 'add' requires <name> and <description>");
        eprintln!("Usage: task-tracker add <name> <description>");
        process::exit(1);
    }

    let mut store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading tasks: {}", e);
        process::exit(1);
    });

    let task = Task::new(store.next_id, args[0].clone(), args[1].clone(), user.id);
    println!("Task added successfully (ID: {})", task.id);

    store.tasks.push(task);
    store.next_id += 1;

    s.save(&store).unwrap_or_else(|e| {
        eprintln!("Error saving tasks: {}", e);
        process::exit(1);
    });
}

fn handle_list(s: &Storage, args: &[String], user: &User) {
    let store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading tasks: {}", e);
        process::exit(1);
    });

    let filter_state = if !args.is_empty() {
        Some(parse_state(&args[0]).unwrap_or_else(|e| {
            eprintln!("Error: {}", e);
            process::exit(1);
        }))
    } else {
        None
    };

    println!(
        "{:<4} {:<20} {:<30} {:<12} {}",
        "ID", "Name", "Description", "State", "Updated"
    );
    println!(
        "---- -------------------- ------------------------------ ------------ -------------------"
    );

    let mut found = false;
    for task in &store.tasks {
        if task.owner != user.id {
            continue;
        }
        if let Some(ref state) = filter_state {
            if task.state != *state {
                continue;
            }
        }
        found = true;
        println!(
            "{:<4} {:<20} {:<30} {:<12} {}",
            task.id,
            truncate(&task.name, 20),
            truncate(&task.description, 30),
            task.state,
            task.updated_at,
        );
    }

    if !found {
        match filter_state {
            Some(state) => println!("No tasks with state {}.", state),
            None => println!("No tasks found."),
        }
    }
}

fn handle_update(s: &Storage, args: &[String], user: &User) {
    if args.len() < 3 {
        eprintln!("Error: 'update' requires <id> <name> <description>");
        eprintln!("Usage: task-tracker update <id> <name> <description>");
        process::exit(1);
    }

    let id: i32 = args[0].parse().unwrap_or_else(|_| {
        eprintln!("Error: invalid ID {:?} - must be a number", args[0]);
        process::exit(1);
    });

    let mut store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading tasks: {}", e);
        process::exit(1);
    });

    // El bloque limita el alcance del préstamo mutable de `store`,
    // permitiendo que `s.save(&store)` tome una referencia inmutable después.
    {
        let (task, _) = s.find_by_id(&mut store, id).unwrap_or_else(|e| {
            eprintln!("Error: {}", e);
            process::exit(1);
        });

        if task.owner != user.id {
            eprintln!("Error: you can only update your own tasks");
            process::exit(1);
        }

        task.name = args[1].clone();
        task.description = args[2].clone();
        task.updated_at = Utc::now().to_rfc3339();
    }

    s.save(&store).unwrap_or_else(|e| {
        eprintln!("Error saving tasks: {}", e);
        process::exit(1);
    });

    println!("Task {} updated successfully.", id);
}

fn handle_delete(s: &Storage, args: &[String], user: &User) {
    if args.is_empty() {
        eprintln!("Error: 'delete' requires <id>");
        eprintln!("Usage: task-tracker delete <id>");
        process::exit(1);
    }

    let id: i32 = args[0].parse().unwrap_or_else(|_| {
        eprintln!("Error: invalid ID {:?} - must be a number", args[0]);
        process::exit(1);
    });

    let mut store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading tasks: {}", e);
        process::exit(1);
    });

    // Extraemos el índice y verificamos propiedad dentro de un bloque
    // para liberar el préstamo mutable antes de llamar a `remove` y `save`.
    let index = {
        let (task, idx) = s.find_by_id(&mut store, id).unwrap_or_else(|e| {
            eprintln!("Error: {}", e);
            process::exit(1);
        });

        if task.owner != user.id {
            eprintln!("Error: you can only delete your own tasks");
            process::exit(1);
        }

        idx
    };

    store.tasks.remove(index);

    s.save(&store).unwrap_or_else(|e| {
        eprintln!("Error saving tasks: {}", e);
        process::exit(1);
    });

    println!("Task {} deleted successfully.", id);
}

fn handle_status(s: &Storage, args: &[String], user: &User) {
    if args.len() < 2 {
        eprintln!("Error: 'status' requires <id> <state>");
        eprintln!("Usage: task-tracker status <id> <state>");
        eprintln!("States: PENDING, IN_PROGRESS, DONE");
        process::exit(1);
    }

    let id: i32 = args[0].parse().unwrap_or_else(|_| {
        eprintln!("Error: invalid ID {:?} - must be a number", args[0]);
        process::exit(1);
    });

    let new_state = parse_state(&args[1]).unwrap_or_else(|e| {
        eprintln!("Error: {}", e);
        process::exit(1);
    });

    let mut store = s.load().unwrap_or_else(|e| {
        eprintln!("Error loading tasks: {}", e);
        process::exit(1);
    });

    {
        let (task, _) = s.find_by_id(&mut store, id).unwrap_or_else(|e| {
            eprintln!("Error: {}", e);
            process::exit(1);
        });

        if task.owner != user.id {
            eprintln!("Error: you can only change the status of your own tasks");
            process::exit(1);
        }

        task.state = new_state.clone();
        task.updated_at = Utc::now().to_rfc3339();
    }

    s.save(&store).unwrap_or_else(|e| {
        eprintln!("Error saving tasks: {}", e);
        process::exit(1);
    });

    println!("Task {} state changed to {}.", id, new_state);
}

/// Ajusta la longitud de una cadena de texto para fines de visualización.
/// Opera a nivel de `chars()` (Unicode) para evitar truncar bytes parciales.
fn truncate(s: &str, max_chars: usize) -> String {
    let chars: Vec<char> = s.chars().collect();
    if chars.len() <= max_chars {
        s.to_string()
    } else {
        let truncated: String = chars[..max_chars - 3].iter().collect();
        format!("{}...", truncated)
    }
}
