# ⚠️ Proyecto en Desuso ⚠️

**Aviso:** Este proyecto ya no está en soporte. El desarrollo se ha trasladado a un nuevo repositorio: [Melodix](https://github.com/keshon/melodix). Por favor, visita el nuevo proyecto para obtener las últimas actualizaciones y soporte.

![# Header](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](./README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](./README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](./README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](./README_JP.md)

# 🎵 Melodix Player — Bot de música para Discord auto-hospedado escrito en Go

Melodix Player es mi proyecto personal que reproduce audio de YouTube y enlaces de transmisión de audio en los canales de voz de Discord.

![Ejemplo de Reproducción](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/demo.gif)

## 🌟 Resumen de Características

### 🎧 Soporte de Reproducción
- 🎶 Pista única añadida por nombre de canción o enlace de YouTube.
- 🎶 Múltiples pistas añadidas mediante múltiples enlaces de YouTube (separados por espacios).
- 🎶 Pistas de listas de reproducción públicas de usuarios.
- 🎶 Pistas de listas de reproducción "MIX".
- 📻 Enlaces de transmisión (por ejemplo, estaciones de radio).

### ⚙️ Características Adicionales
- 🌐 Operación en múltiples servidores de Discord (gestión de gremios).
- 📜 Acceso al historial de pistas reproducidas previamente con opciones de clasificación.
- 💾 Descarga de pistas de YouTube como archivos mp3 para almacenamiento en caché.
- 🎼 Carga lateral de archivos de audio mp3.
- 🎬 Carga lateral de archivos de video con extracción de audio como archivos mp3.
- 🔄 Soporte de reanudación automática de reproducción para interrupciones de conexión.
- 🛠️ Soporte de API REST (limitado por el momento).

### ⚠️ Limitaciones Actuales
- 🚫 El bot no puede reproducir transmisiones de YouTube.
- ⏸️ El soporte de reanudación automática de reproducción crea pausas notables.
- ⏩ A veces, la velocidad de reproducción es ligeramente más rápida de lo previsto.
- 🐞 No está libre de errores.

## 🚀 Prueba Melodix Player

Puedes probar Melodix de dos maneras:
- 🖥️ Descarga [binarios compilados](https://github.com/keshon/melodix-player/releases) (disponibles solo para Windows). Asegúrate de tener FFMPEG instalado en tu sistema y añadido a la variable PATH global (o especifica la ruta a FFMPEG directamente en el archivo de configuración `.env`). Sigue la sección "Crear bot en el Portal de Desarrolladores de Discord" para configurar el bot en Discord.

- 🎙️ Únete al [Servidor Oficial de Discord](https://discord.gg/NVtdTka8ZT) y usa los canales de voz y `#bot-spam`.

## 📝 Comandos Disponibles en Discord

Melodix Player soporta varios comandos con respectivos alias (si aplica). Algunos comandos requieren parámetros adicionales.

### ▶️ Comandos de Reproducción
- `!play [título|url|stream|id]` (alias: `!p ..`, `!> ..`) — Parámetros: nombre de la canción, URL de YouTube, URL de transmisión de audio, ID del historial.
- `!skip` (alias: `!next`, `!>>`) — Saltar a la siguiente pista en la cola.
- `!pause` (alias: `!!`) — Pausar la reproducción.
- `!resume` (alias: `!r`, `!!>`) — Reanudar la reproducción pausada o iniciar la reproducción si se añadió una pista mediante `!add ..`.
- `!stop` (alias: `!x`) — Detener la reproducción, limpiar la cola y salir del canal de voz.

### 📋 Comandos de Cola
- `!add [título|url|stream|id]` (alias: `!a`, `!+`) — Parámetros: nombre de la canción, URL de YouTube, URL de transmisión de audio, ID del historial (igual que para `!play ..`).
- `!list` (alias: `!queue`, `!l`, `!q`) — Mostrar la cola de canciones actual.

### 📚 Comandos de Historial
- `!history` (alias: `!time`, `!t`) — Mostrar el historial de pistas reproducidas recientemente. Cada pista en el historial tiene un ID único para reproducción/colocación en la cola.
- `!history count` (alias: `!time count`, `!t count`) — Ordenar el historial por recuento de reproducciones.
- `!history duration` (alias: `!time duration`, `!t duration`) — Ordenar el historial por duración de las pistas.

### ℹ️ Comandos de Información
- `!help` (alias: `!h`, `!?`) — Mostrar hoja de trucos de ayuda.
- `!help play` — Información adicional sobre comandos de reproducción.
- `!help queue` — Información adicional sobre comandos de cola.
- `!about` (alias: `!v`) — Mostrar versión (fecha de compilación) y enlaces relacionados.
- `whoami` — Enviar información del usuario al log. Necesario para configurar el superadmin en el archivo `.env`.

### 💾 Comandos de Caché y Carga Lateral
Estos comandos solo están disponibles para superadmins (propietarios del servidor host).
- `!curl [URL de YouTube]` — Descargar como archivo mp3 para uso posterior.
- `!cached` — Mostrar archivos actualmente en caché (de la carpeta `cached`). Cada servidor opera sus propios archivos.
- `!cached sync` — Sincronizar archivos mp3 añadidos manualmente a la carpeta `cached`.
- `!uploaded` — Mostrar videoclips subidos en la carpeta `uploaded`.
- `!uploaded extract` — Extraer archivos mp3 de videoclips y almacenarlos en la carpeta `cached`.

### 🔧 Comandos de Administración
- `!register` — Habilitar la escucha de comandos de Melodix (ejecutar una vez por cada nuevo servidor de Discord).
- `!unregister` — Deshabilitar la escucha de comandos.
- `melodix-prefix` — Mostrar el prefijo actual (`!` por defecto, ver archivo `.env`).
- `melodix-prefix-update "[nuevo_prefijo]"` — Establecer un prefijo personalizado para un gremio para evitar colisiones con otros bots.
- `melodix-prefix-reset` — Volver al prefijo por defecto establecido en el archivo `.env`.

### 💡 Ejemplos de Uso de Comandos
Para usar el comando `play`, proporciona un título de video de YouTube, URL o ID del historial:
```
!play Never Gonna Give You Up
!p https://www.youtube.com/watch?v=dQw4w9WgXcQ
!> 5  (asumiendo que 5 es un ID de `!history`)
```
Para añadir una canción a la cola, usa:
```
!add Never Gonna Give You Up
!resume
```

## 🔧 Cómo Configurar el Bot

### 🔗 Crear un Bot en el Portal de Desarrolladores de Discord
Para añadir Melodix a un servidor de Discord, sigue estos pasos:

1. Crea una aplicación en el [Portal de Desarrolladores de Discord](https://discord.com/developers/applications) y obtén el `APPLICATION_ID` (en la sección General).
2. En la sección Bot, habilita `PRESENCE INTENT`, `SERVER MEMBERS INTENT`, y `MESSAGE CONTENT INTENT`.
3. Usa el siguiente enlace para autorizar el bot: `discord.com/oauth2/authorize?client_id=YOUR_APPLICATION_ID&scope=bot&permissions=36727824`
   - Reemplaza `YOUR_APPLICATION_ID` con el ID de la aplicación de tu bot del paso 1.
4. Selecciona un servidor y haz clic en "Autorizar".
5. Concede los permisos necesarios para que Melodix funcione correctamente (acceso a canales de texto y voz).

Después de añadir el bot, compílalo desde los fuentes o descarga [binarios compilados](https://github.com/keshon/melodix-player/releases). Las instrucciones de despliegue con Docker están disponibles en `docker/README.md`.

### 🛠️ Compilar Melodix desde los Fuentes
Este proyecto está escrito en Go, así que asegúrate de que tu entorno esté listo. Usa los scripts proporcionados para compilar Melodix Player desde los fuentes:
- `bash-and-run.bat` (o `.sh` para Linux): Compilar la versión de depuración y ejecutar.
- `build-release.bat` (o `.sh` para Linux): Compilar la versión de lanzamiento.
- `assemble-dist.bat`: Compilar la versión de lanzamiento y ensamblarla como paquete de distribución (solo Windows).

Renombra `.env.example` a `.env` y guarda tu Token del Bot de Discord en la variable `DISCORD_BOT_TOKEN`. Instala [FFMPEG](https://ffmpeg.org/) (solo se soportan versiones recientes). Si usas un FFMPEG portátil, especifica la ruta en `DCA_FFMPEG_BINARY_PATH

` en el archivo `.env`.

### 🐳 Despliegue con Docker
Para el despliegue con Docker, consulta `docker/README.md` para instrucciones específicas.

## 🌐 API REST
Melodix Player proporciona varias rutas de API, sujetas a cambios.

### Rutas de Gremios
- `GET /guild/ids`: Recuperar IDs de gremios activos.
- `GET /guild/playing`: Obtener información sobre la pista actualmente reproducida en cada gremio activo.

### Rutas de Historial
- `GET /history`: Acceder al historial general de pistas reproducidas.
- `GET /history/:guild_id`: Obtener el historial de pistas reproducidas para un gremio específico.

### Rutas de Avatar
- `GET /avatar`: Listar imágenes disponibles en la carpeta de avatares.
- `GET /avatar/random`: Obtener una imagen aleatoria de la carpeta de avatares.

### Rutas de Log
- `GET /log`: Mostrar el log actual.
- `GET /log/clear`: Limpiar el log.
- `GET /log/download`: Descargar el log como archivo.

## 🆘 Soporte
Para cualquier pregunta, obtén soporte en el [Servidor Oficial de Discord](https://discord.gg/NVtdTka8ZT).

## 🏆 Agradecimientos
Me inspiré en [Muzikas](https://github.com/FabijanZulj/Muzikas), un bot de Discord fácil de usar creado por Fabijan Zulj.

## 📜 Licencia
Melodix está licenciado bajo la [Licencia MIT](https://opensource.org/licenses/MIT).