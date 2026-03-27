use serde::{Serialize, Deserialize};
use std::fs;
use std::io;

use crate::task::Task;
use crate::user::User;

/// Estructura que representa el esquema del archivo JSON.
/// `next_id` mantiene la integridad de los identificadores únicos.
#[derive(Serialize, Deserialize)]
pub struct TaskStore {
    #[serde(rename = "next_id")]
    pub next_id: i32,
    #[serde(default = "default_next_user_id")]
    pub next_user_id: i32,
    pub tasks: Vec<Task>,
    #[serde(default)]
    pub users: Vec<User>,
}

fn default_next_user_id() -> i32 {
    1
}

/// Capa de abstracción para el acceso a datos.
pub struct Storage {
    pub filepath: String,
}

impl Storage {
    pub fn new(filepath: String) -> Self {
        Storage { filepath }
    }

    /// Carga el contenido del archivo JSON y lo deserializa en un `TaskStore`.
    /// Retorna un error dinámico (`Box<dyn std::error::Error>`) si el archivo es corrupto 
    /// o no es accesible. Si el archivo no existe, inicializa una estructura vacía.
    pub fn load(&self) -> Result<TaskStore, Box<dyn std::error::Error>> {
        let store = TaskStore {
            next_id: 1,
            next_user_id: 1,
            tasks: Vec::new(),
            users: Vec::new(),
        };

        // Uso de pattern matching para diferenciar un archivo inexistente de un error de lectura.
        let data = match fs::read_to_string(&self.filepath) {
            Ok(content) => content,
            Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(store),
            Err(e) => return Err(format!("reading file: {}", e).into()),
        };

        let parsed_store: TaskStore = serde_json::from_str(&data)
            .map_err(|e| format!("parsing JSON: {}", e))?;

        Ok(parsed_store)
    }

    /// Serializa el `TaskStore` a formato JSON con sangría (pretty-print) y lo escribe en disco.
    pub fn save(&self, store: &TaskStore) -> Result<(), Box<dyn std::error::Error>> {
        let data = serde_json::to_string_pretty(store)
            .map_err(|e| format!("encoding JSON: {}", e))?;
        fs::write(&self.filepath, data)
            .map_err(|e| format!("writing file: {}", e).into())
    }

    /// Busca una tarea por su ID dentro del almacén.
    /// Retorna una referencia mutable a la tarea encontrada y su índice en el vector.
    /// El parámetro de vida `'a` vincula la validez de la referencia de salida con el almacén de entrada.
    pub fn find_by_id<'a>(
        &self,
        store: &'a mut TaskStore,
        id: i32,
    ) -> Result<(&'a mut Task, usize), Box<dyn std::error::Error>> {
        let index = store.tasks.iter().position(|t| t.id == id);

        match index {
            Some(i) => Ok((&mut store.tasks[i], i)),
            None => Err(format!("task with ID {} not found", id).into()),
        }
    }

    /// Busca un usuario por su nombre de usuario dentro del almacén.
    pub fn find_user_by_username<'a>(
        &self,
        store: &'a TaskStore,
        username: &str,
    ) -> Result<&'a User, Box<dyn std::error::Error>> {
        for user in &store.users {
            if user.username == username {
                return Ok(user);
            }
        }
        Err(format!("user {:?} not found", username).into())
    }
}