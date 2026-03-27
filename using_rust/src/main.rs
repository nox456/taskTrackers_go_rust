use std::env;
use std::process;
use chrono::Utc;

/// Punto de entrada del programa.
/// Orquesta la lectura de argumentos del sistema y la ejecución de los controladores correspondientes.
fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        print_usage();
        process::exit(1);
    }

    let storage = Storage::new(DATA_FILE.to_string());
    let command = &args[1];
    let cmd_args = &args[2..]; // Slice de argumentos restantes para los handlers específicos.

    match command.as_str() {
        "add" => handle_add(&storage, cmd_args),
        "list" => handle_list(&storage, cmd_args),
        "update" => handle_update(&storage, cmd_args),
        "delete" => handle_delete(&storage, cmd_args),
        "status" => handle_status(&storage, cmd_args),
        _ => {
            eprintln!("Unknown command: {}", command);
            print_usage();
            process::exit(1);
        }
    }
}

/// Handler para el borrado de tareas.
/// Utiliza `s.find_by_id` para localizar el índice y aplica la mutación directamente sobre el vector.
fn handle_delete(s: &Storage, args: &[String]) {
    if args.is_empty() {
        eprintln!("Error: 'delete' requires <id>");
        process::exit(1);
    }

    // El método parse() convierte el string a i32 de forma segura.
    let id: i32 = args[0].parse().expect("Invalid ID");
    let mut store = s.load().expect("Load error");

    match s.find_by_id(&mut store, id) {
        Ok((_, index)) => {
            store.tasks.remove(index);
            s.save(&store).expect("Save error");
            println!("Task {} deleted successfully.", id);
        }
        Err(e) => eprintln!("Error: {}", e),
    }
}

/// Ajusta la longitud de una cadena de texto para fines de visualización.
/// Esta función opera a nivel de `chars()` (Unicode) para evitar truncar bytes parciales, 
/// garantizando la integridad de caracteres especiales.
fn truncate(s: &str, max_chars: usize) -> String {
    let chars: Vec<char> = s.chars().collect();
    if chars.len() <= max_chars {
        s.to_string()
    } else {
        // Se sustraen 3 para acomodar los puntos suspensivos (...)
        let truncated: String = chars[..max_chars - 3].iter().collect();
        format!("{}...", truncated)
    }
}