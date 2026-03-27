use serde::{Serialize, Deserialize};
use chrono::{Utc};
use std::fmt;

/// Representa los estados posibles de una tarea.
/// El atributo `#[serde(rename_all = "...")]` garantiza la compatibilidad con el 
/// formato de datos del proyecto en Go, serializando los variantes como strings en mayúsculas.
#[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
#[serde(rename_all = "SCREAMING_SNAKE_CASE")]
pub enum State {
    Pending,
    InProgress,
    Done,
}

/// Implementación del trait `Display` para permitir la representación textual del estado.
/// Reemplaza la necesidad de funciones manuales de "string conversion" usadas en otros lenguajes.
impl fmt::Display for State {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let s = match self {
            State::Pending => "PENDING",
            State::InProgress => "IN_PROGRESS",
            State::Done => "DONE",
        };
        write!(f, "{}", s)
    }
}

/// Estructura principal que define una Tarea.
/// Se utiliza `derive` para implementar automáticamente la serialización (Serde)
/// y capacidades de clonación profunda (Clone).
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Task {
    pub id: i32,
    pub name: String,
    pub description: String,
    pub state: State,
    pub owner: i32,
    pub created_at: String,
    pub updated_at: String,
}

/// Convierte una cadena de texto en una variante del enum `State`.
/// Retorna `Result` para manejar de forma segura entradas de usuario no válidas,
/// evitando fallos en tiempo de ejecución.
pub fn parse_state(s: &str) -> Result<State, String> {
    match s.to_uppercase().as_str() {
        "PENDING" => Ok(State::Pending),
        "IN_PROGRESS" | "IN PROGRESS" | "INPROGRESS" => Ok(State::InProgress),
        "DONE" => Ok(State::Done),
        _ => Err(format!(
            "invalid state {:?}: must be PENDING, IN_PROGRESS, or DONE",
            s
        )),
    }
}

impl Task {
    /// Constructor de la estructura Task.
    /// Inicializa las marcas de tiempo en formato ISO 8601 (RFC3339) utilizando la crate `chrono`.
    pub fn new(id: i32, name: String, description: String, owner_id: i32) -> Self {
        let now = Utc::now().to_rfc3339();
        Task {
            id,
            name,
            description,
            state: State::Pending,
            owner: owner_id,
            created_at: now.clone(),
            updated_at: now,
        }
    }
}