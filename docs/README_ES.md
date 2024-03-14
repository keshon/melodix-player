![# Header](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](/docs/README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](/docs/README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](/docs/README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](/docs/README_JP.md)

# Melodix Player

Melodix Player es un bot de música para Discord que hace lo mejor posible, incluso en presencia de errores de conexión.

## Descripción General de Funciones

El bot tiene como objetivo ser un reproductor de música fácil de usar pero potente. Sus principales objetivos incluyen:

- Reproducción de pistas individuales o múltiples, así como listas de reproducción de YouTube, agregadas por título o URL.
- Reproducción de transmisiones de radio agregadas mediante URL.
- Acceso al historial de pistas reproducidas anteriormente con opciones de ordenación según recuentos de reproducción o duración.
- Manejo de interrupciones de reproducción debido a fallas de red; Melodix intentará reanudar la reproducción.
- API Rest expuesta para realizar varias tareas fuera de los comandos de Discord.
- Operación en varios servidores de Discord.

![Ejemplo de Reproducción](https://github.com/keshon/melodix-player/blob/master/assets/demo.gif)

## Descarga del Binario

Los binarios (solo para Windows) están disponibles en la [página de lanzamientos](https://github.com/keshon/melodix-player/releases). Se recomienda compilar los binarios desde el código fuente para obtener la última versión.

## Comandos de Discord

Melodix Player admite varios comandos con sus respectivos alias para controlar la reproducción de música. Algunos comandos requieren parámetros adicionales:

**Comandos y Alias**:
- `play` (`p`, `>`) — Parámetros: URL de video de YouTube, ID de historial, título de la pista o enlace de transmisión válido.
- `skip` (`next`, `ff`, `>>`)
- `pause` (`!`)
- `resume` (`r`,`!>`)
- `stop` (`x`)
- `add` (`a`, `+`) — Parámetros: URL de video de YouTube o ID de historial, título de la pista o enlace de transmisión válido.
- `list` (`queue`, `l`, `q`)
- `history` (`time`, `t`) — Parámetros: `duración` o `recuento`
- `help` (`h`, `?`)
- `about` (`v`)
- `register`
- `unregister`

Los comandos deben tener un prefijo de `!` por defecto. Por ejemplo, `!play`, `!>>`, y así sucesivamente.

### Ejemplos
Para usar el comando `play`, proporciona el título de un video de YouTube, URL o un ID de historial como parámetro, por ejemplo:
`!play Never Gonna Give You Up` 
o 
`!p https://www.youtube.com/watch?v=dQw4w9WgXcQ` 
o 
`!> 5` (suponiendo que `5` es un ID que se puede ver en el historial: `!history`)

De manera similar, para agregar una canción a la cola, utiliza un enfoque similar:
`!add Never Gonna Give You Up` 
`!resume` (para comenzar a reproducir)

## Agregar el Bot a un Servidor de Discord

Para agregar Melodix a tu servidor de Discord:

1. Crea un bot en el [Portal de Desarrolladores de Discord](https://discord.com/developers/applications) y obtén el CLIENT_ID del bot.
2. Utiliza el siguiente enlace: `discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID_HERE&scope=bot&permissions=36727824`
   - Reemplaza `YOUR_CLIENT_ID_HERE` con el CLIENT_ID de tu bot obtenido en el paso 1.
3. La página de autorización de Discord se abrirá en tu navegador, permitiéndote seleccionar un servidor.
4. Elige el servidor donde deseas agregar Melodix y haz clic en "Autorizar".
5. Si se solicita, completa la verificación reCAPTCHA.
6. Concede a Melodix los permisos necesarios para que funcione correctamente.
7. Haz clic en "Autorizar" para agregar Melodix a tu servidor.

Una vez que el bot esté agregado, continúa con la construcción real del bot.

## Acceso a la API y Rutas

Melodix Player proporciona varias rutas para diferentes funcionalidades:

### Rutas del Servidor

- `GET /guild/ids`: Obtiene los IDs activos de los servidores.
- `GET /guild/playing`: Obtiene información sobre la pista que se está reproduciendo actualmente en cada servidor activo.

### Rutas del Historial

- `GET /history`: Accede al historial general de pistas reproducidas.
- `GET /history/:guild_id`: Obtiene el historial de pistas reproducidas para un servidor específico.

### Rutas de Avatar

- `GET /avatar`: Lista las imágenes disponibles en la carpeta de avatares.
- `GET /avatar/random`: Obtiene una imagen aleatoria de la carpeta de avatares.

### Rutas de Registro

- `GET /log`: Muestra el registro actual.
- `GET /log/clear`: Borra el registro.
- `GET /log/download`: Descarga el registro como un archivo.

## Construcción desde el Código Fuente

Este proyecto está escrito en el lenguaje Go, lo que permite que se ejecute en un *servidor* o como un programa *local*.

**Uso Local**
Se proporcionan varios scripts para construir Melodix Player desde el código fuente:
- `bash-and-run.bat` (o `.sh` para Linux): Construye la versión de depuración y la ejecuta.
- `build-release.bat` (o `.sh` para Linux): Construye la versión de lanzamiento.
- `assemble-dist.bat`: Construye la versión de lanzamiento y la ensambla como un paquete de distribución (solo Windows, el empaquetador UPX se descargará durante el proceso).

Para el uso local, ejecuta estos scripts para tu sistema operativo y renombra `.env.example` a `.env`, almacenando tu Token de Bot de Discord en la variable `DISCORD_BOT_TOKEN`. Instala [FFMPEG](https://ffmpeg.org/) (solo se admite la versión más reciente). Si tu instalación de FFMPEG es portátil, especifica la ruta en la variable `DCA_FFMPEG_BINARY_PATH`.

**Uso en Servidor**
Para construir y desplegar el bot en un entorno Docker, consulta el archivo `docker/README.md` para obtener instrucciones específicas.

Una vez que el archivo binario está construido, el archivo `.env` está lleno y el bot está agregado a tu servidor, Melodix está listo para funcionar.

## ¿Dónde Obtener Soporte
Si tienes alguna pregunta, puedes preguntarme en mi [servidor de Discord](https://discord.gg/NVtdTka8ZT) para obtener soporte. Ten en cuenta que no hay comunidad en absoluto, solo yo.

## Reconocimientos

Me inspiré en [Muzikas](https://github.com/FabijanZulj/Muzikas), un bot de Discord fácil de usar creado por Fabijan Zulj.

Como resultado del desarrollo de Melodix, nació un nuevo proyecto: [Discord Bot Boilerplate](https://github.com/keshon/discord-bot-boilerplate) — un marco para construir bots de Discord.

## Licencia

Melodix tiene licencia bajo la [Licencia MIT](https://opensource.org/licenses/MIT).