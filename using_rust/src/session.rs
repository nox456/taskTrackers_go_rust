use std::fs;
use std::io;

use crate::user::{User, hash_sha256};

/// Archivo de sesión que almacena el hash SHA256 del ID del usuario autenticado.
const SESSION_FILE: &str = ".session";

/// Guarda la sesión del usuario escribiendo el hash SHA256 de su ID en el archivo de sesión.
pub fn save_session(user_id: i32) -> Result<(), Box<dyn std::error::Error>> {
    let hash = hash_sha256(&user_id.to_string());
    fs::write(SESSION_FILE, hash)?;
    Ok(())
}

/// Carga la sesión activa comparando el hash almacenado con los IDs de los usuarios registrados.
/// Retorna un `User` clonado si encuentra una coincidencia.
pub fn load_session(users: &[User]) -> Result<User, Box<dyn std::error::Error>> {
    let data = match fs::read_to_string(SESSION_FILE) {
        Ok(content) => content,
        Err(e) if e.kind() == io::ErrorKind::NotFound => {
            return Err("no active session. Please login first".into());
        }
        Err(e) => return Err(format!("reading session: {}", e).into()),
    };

    let session_hash = data.trim();

    for user in users {
        let user_hash = hash_sha256(&user.id.to_string());
        if user_hash == session_hash {
            return Ok(user.clone());
        }
    }

    Err("session expired or invalid. Please login again".into())
}

/// Elimina el archivo de sesión. No retorna error si el archivo no existe.
pub fn clear_session() -> Result<(), Box<dyn std::error::Error>> {
    match fs::remove_file(SESSION_FILE) {
        Ok(()) => Ok(()),
        Err(e) if e.kind() == io::ErrorKind::NotFound => Ok(()),
        Err(e) => Err(format!("removing session: {}", e).into()),
    }
}
