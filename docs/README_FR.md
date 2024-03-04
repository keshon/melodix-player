![# Header](https://github.com/keshon/melodix-player/blob/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](/docs/README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](/docs/README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](/docs/README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](/docs/README_JP.md)

# Melodix Player

Melodix Player est un bot musical Discord qui fait de son mieux, même en cas d'erreurs de connexion.

## Aperçu des fonctionnalités

Le bot vise à être un lecteur musical facile à utiliser mais puissant. Ses objectifs clés comprennent :

- Lecture de pistes individuelles/multiples ou de listes de lecture depuis YouTube, ajoutées par titre ou URL.
- Lecture de flux radio ajoutés via une URL.
- Accès à l'historique des pistes précédemment jouées avec des options de tri pour le nombre de lectures ou la durée.
- Gestion des interruptions de lecture dues à des échecs de réseau - Melodix tentera de reprendre la lecture.
- API Rest exposée pour effectuer diverses tâches en dehors des commandes Discord.
- Fonctionnement sur plusieurs serveurs Discord.

![Exemple de lecture](https://github.com/keshon/melodix-player/blob/master/assets/playing.jpg)

## Téléchargement du binaire

Les binaires (uniquement pour Windows) sont disponibles sur la [page des versions](https://github.com/keshon/melodix-player/releases). Il est recommandé de construire les binaires à partir des sources pour la dernière version.

## Commandes Discord

Melodix Player prend en charge diverses commandes avec leurs alias respectifs pour contrôler la lecture de musique. Certaines commandes nécessitent des paramètres supplémentaires :

**Commandes & Alias**:
- `play` (`p`, `>`) — Paramètres : URL vidéo YouTube, ID d'historique, titre de la piste ou lien de flux valide
- `skip` (`next`, `ff`, `>>`)
- `pause` (`!`)
- `resume` (`r`,`!>`)
- `stop` (`x`)
- `add` (`a`, `+`) — Paramètres : URL vidéo YouTube ou ID d'historique, titre de la piste ou lien de flux valide
- `list` (`queue`, `l`, `q`)
- `history` (`time`, `t`) — Paramètres : `durée` ou `nombre`
- `help` (`h`, `?`)
- `about` (`v`)
- `register`
- `unregister`

Les commandes doivent être préfixées par `!` par défaut. Par exemple, `!play`, `!>>`, et ainsi de suite.

### Exemples
Pour utiliser la commande `play`, fournissez un titre vidéo YouTube, une URL ou un ID d'historique en tant que paramètre, par exemple :
`!play Never Gonna Give You Up` 
ou 
`!p https://www.youtube.com/watch?v=dQw4w9WgXcQ` 
ou 
`!> 5` (en supposant que `5` est un ID visible dans l'historique : `!history`)

Pour ajouter une chanson à la file d'attente, utilisez une approche similaire :
`!add Never Gonna Give You Up` 
`!resume` (pour commencer la lecture)

## Ajout du bot à un serveur Discord

Pour ajouter Melodix à votre serveur Discord :

1. Créez un bot sur le [Portail des développeurs Discord](https://discord.com/developers/applications) et obtenez l'ID_CLIENT du bot.
2. Utilisez le lien suivant : `discord.com/oauth2/authorize?client_id=VOTRE_ID_CLIENT_ICI&scope=bot&permissions=36727824`
   - Remplacez `VOTRE_ID_CLIENT_ICI` par l'ID_CLIENT de votre bot à partir de l'étape 1.
3. La page d'autorisation Discord s'ouvrira dans votre navigateur, vous permettant de sélectionner un serveur.
4. Choisissez le serveur où vous souhaitez ajouter Melodix et cliquez sur "Autoriser".
5. Si on vous le demande, effectuez la vérification reCAPTCHA.
6. Accordez à Melodix les permissions nécessaires pour son bon fonctionnement.
7. Cliquez sur "Autoriser" pour ajouter Melodix à votre serveur.

Une fois le bot ajouté, procédez à la construction réelle du bot.

## Accès à l'API et Routes

Melodix Player fournit diverses routes pour différentes fonctionnalités :

### Routes de la guilde

- `GET /guild/ids` : Récupérer les ID de guilde actifs.
- `GET /guild/playing` : Obtenir des informations sur la piste actuellement en cours de lecture dans chaque guilde active.

### Routes de l'historique

- `GET /history` : Accéder à l'historique global des pistes jouées.
- `GET /history/:guild_id` : Récupérer l'historique des pistes jouées pour une guilde spécifique.

### Routes de l'avatar

- `GET /avatar` : Liste des images disponibles dans le dossier avatar.
- `GET /avatar/random` : Récupérer une image aléatoire dans le dossier avatar.

### Routes du journal

- `GET /log` : Afficher le journal actuel.
- `GET /log/clear` : Effacer le journal.
- `GET /log/download` : Télécharger le journal sous forme de fichier.

## Construction à partir des sources

Ce projet est écrit en langage Go, ce qui lui permet de s'exécuter sur un *serveur* ou en tant que programme *local*.

**Utilisation locale**
Plusieurs scripts sont fournis pour construire Melodix Player à partir des sources :
- `bash-and-run.bat` (ou `.sh` pour Linux) : Construire la version de débogage et l'exécuter.
- `build-release.bat` (ou `.sh` pour Linux) : Construire la version de production.
- `assemble-dist.bat` : Construire la version de production et l'assembler comme un package de distribution (uniquement pour Windows, le packager UPX sera téléchargé pendant le processus).

Pour une utilisation locale, exécutez ces scripts pour votre système d'exploitation et renommez `.env.example` en `.env`, enregistrant votre jeton Discord Bot dans la variable `DISCORD_BOT_TOKEN`. Installez [FFMPEG](https://ffmpeg.org/) (seule la version récente est prise en charge). Si votre installation FFMPEG est portable, spécifiez le chemin dans la variable `DCA_FFMPEG_BINARY_PATH`.

**Utilisation sur un serveur**
Pour construire et déployer le bot dans un environnement Docker, consultez le fichier `docker/README.md` pour des instructions spécifiques.

Une fois le fichier binaire construit, le fichier `.env` rempli et le bot ajouté à votre serveur, Melodix est prêt à fonctionner.

## Où obtenir du support
Si vous avez des questions, vous pouvez me les poser sur [mon serveur Discord](https://discord.gg/NVtdTka8ZT) pour obtenir de l'aide. Gardez à l'esprit qu'il n'y a aucune communauté - juste moi.

## Remerciements

Je me suis inspiré de [Muzikas](https://github.com/FabijanZulj/Muzikas), un bot Discord convivial créé par Fabijan Zulj.

À la suite du développement de Melodix, un nouveau projet est né : [Discord Bot Boilerplate](https://github.com/keshon/discord-bot-boilerplate) — un cadre pour construire des bots Discord.

## Licence

Melodix est sous licence [MIT](https://opensource.org/licenses/MIT).