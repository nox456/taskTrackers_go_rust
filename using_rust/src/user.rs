use serde::{Serialize, Deserialize};
use sha2::{Sha256, Digest};

use crate::task::Task;

/// Representa un usuario autenticado en el sistema.
/// El campo `tasks` utiliza `#[serde(skip)]` para excluirlo de la serialización JSON.
/// Es un campo calculado que se llena en tiempo de ejecución filtrando tareas por propietario.
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct User {
    pub id: i32,
    pub username: String,
    pub password: String,
    #[serde(skip)]
    pub tasks: Vec<Task>,
}

impl User {
    pub fn new(id: i32, username: String, password: String) -> Self {
        User {
            id,
            username,
            password: hash_sha256(&password),
            tasks: Vec::new(),
        }
    }
}

/// Genera un hash SHA256 de la cadena proporcionada y lo retorna como string hexadecimal.
/// Se utiliza tanto para hashear contraseñas como para la sesión del usuario.
pub fn hash_sha256(s: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(s.as_bytes());
    format!("{:x}", hasher.finalize())
}
