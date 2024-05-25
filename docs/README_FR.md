![# En-tête](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](./README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](./README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](./README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](./README_JP.md)

# 🎵 Melodix Player — Bot musical Discord auto-hébergé écrit en Go

Melodix Player est mon projet personnel qui diffuse de l'audio depuis YouTube et des liens de diffusion audio vers les salons vocaux Discord.

![Exemple de lecture](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/demo.gif)

## 🌟 Aperçu des fonctionnalités

### 🎧 Prise en charge de la lecture
- 🎶 Ajout d'un seul morceau par titre ou lien YouTube.
- 🎶 Ajout de plusieurs morceaux via plusieurs liens YouTube (séparés par des espaces).
- 🎶 Morceaux de listes de lecture publiques d'utilisateurs.
- 🎶 Morceaux de listes de lecture "MIX".
- 📻 Liens de diffusion (par exemple, stations de radio).

### ⚙️ Fonctionnalités supplémentaires
- 🌐 Fonctionnement sur plusieurs serveurs Discord (gestion des guildes).
- 📜 Accès à l'historique des morceaux précédemment joués avec options de tri.
- 💾 Téléchargement de morceaux depuis YouTube au format mp3 pour mise en cache.
- 🎼 Chargement latéral de fichiers audio mp3.
- 🎬 Chargement latéral de fichiers vidéo avec extraction audio au format mp3.
- 🔄 Prise en charge de la reprise automatique de la lecture en cas d'interruption de connexion.
- 🛠️ Prise en charge de l'API REST (limitée pour le moment).

### ⚠️ Limitations actuelles
- 🚫 Le bot ne peut pas diffuser en continu depuis YouTube.
- ⏸️ La reprise automatique de la lecture crée des pauses perceptibles.
- ⏩ Parfois, la vitesse de lecture est légèrement plus rapide que prévu.
- 🐞 Il n'est pas exempt de bugs.

## 🚀 Essayez Melodix Player

Vous pouvez tester Melodix de deux manières :
- 🖥️ Téléchargez les [binaires compilés](https://github.com/keshon/melodix-player/releases) (disponibles uniquement pour Windows). Assurez-vous que FFMPEG est installé sur votre système et ajouté à la variable d'environnement PATH globale (ou spécifiez le chemin vers FFMPEG directement dans le fichier de configuration `.env`). Suivez la section "Créer un bot dans le portail des développeurs Discord" pour configurer le bot dans Discord.

- 🎙️ Rejoignez le [serveur Discord officiel](https://discord.gg/NVtdTka8ZT) et utilisez les canaux vocaux et `#bot-spam`.

## 📝 Commandes Discord disponibles

Melodix Player prend en charge diverses commandes avec des alias respectifs (si applicable). Certaines commandes nécessitent des paramètres supplémentaires.

### ▶️ Commandes de lecture
- `!play [titre|url|flux|id]` (alias : `!p ..`, `!> ..`) — Paramètres : nom du morceau, URL YouTube, URL de diffusion audio, ID de l'historique.
- `!skip` (alias : `!next`, `!>>`) — Passer au morceau suivant dans la file d'attente.
- `!pause` (alias : `!!`) — Mettre la lecture en pause.
- `!resume` (alias : `!r`, `!!>`) — Reprendre la lecture en pause ou démarrer la lecture si un morceau a été ajouté via `!add ..`.
- `!stop` (alias : `!x`) — Arrêter la lecture, vider la file d'attente et quitter le salon vocal.

### 📋 Commandes de file d'attente
- `!add [titre|url|flux|id]` (alias : `!a`, `!+`) — Paramètres : nom du morceau, URL YouTube, URL de diffusion audio, ID de l'historique (identique à celui de `!play ..`).
- `!list` (alias : `!queue`, `!l`, `!q`) — Afficher la file d'attente actuelle des morceaux.

### 📚 Commandes d'historique
- `!history` (alias : `!time`, `!t`) — Afficher l'historique des morceaux récemment joués. Chaque morceau dans l'historique a un ID unique pour la lecture/file d'attente.
- `!history count` (alias : `!time count`, `!t count`) — Trier l'historique par nombre de lectures.
- `!history duration` (alias : `!time duration`, `!t duration`) — Trier l'historique par durée des morceaux.

### ℹ️ Commandes d'informations
- `!help` (alias : `!h`, `!?`) — Afficher une aide sous forme de mémo.
- `!help play` — Informations supplémentaires sur les commandes de lecture.
- `!help queue` — Informations supplémentaires sur les commandes de file d'attente.
- `!about` (alias : `!v`) — Afficher la version (date de construction) et les liens associés.
- `whoami` — Envoyer des informations liées à l'utilisateur dans le journal. Nécessaire pour configurer le superadmin dans le fichier `.env`.

### 💾 Commandes de mise en cache et de chargement latéral
Ces commandes sont uniquement disponibles pour les superadmins (propriétaires de serveur hôte).
- `!curl [URL YouTube]` — Télécharger sous forme de fichier mp3 pour une utilisation ultérieure.
- `!cached` — Afficher les fichiers actuellement mis en cache (du répertoire `cached`). Chaque serveur opère ses propres fichiers.
- `!cached sync` — Synchroniser les fichiers mp3 ajoutés manuellement dans le répertoire `cached`.
- `!uploaded` — Afficher les clips vidéo téléchargés dans le répertoire `uploaded`.
- `!uploaded extract` — Extraire les fichiers mp3 des clips vidéo et les stocker dans le répertoire `cached`.

### 🔧 Commandes d'administration
- `!register` — Activer l'écoute des commandes Melodix (à exécuter une fois pour chaque nouveau serveur Discord).
- `!unregister` — Désactiver l'écoute des commandes.
- `melodix-prefix` — Afficher le préfixe actuel (`!` par défaut, voir le fichier `.env`).
- `melodix-prefix-update "[new_prefix]"` — Définir un préfixe personnalisé pour une guilde afin d'éviter les collisions avec d'autres bots.
- `melodix-prefix-reset` — Revenir au préfixe par défaut défini dans le fichier `.env`.

### 💡 Exemples d'utilisation des commandes
Pour utiliser la commande `play`, fournissez le titre d'une vidéo YouTube, son URL ou son ID d'historique :
```
!play Never Gonna Give You Up
!p https://www.youtube.com/watch?v=dQw4w9WgXcQ
!> 5  (en supposant que 5 est un ID de l'historique à partir de `!history`)
```
Pour ajouter un morceau à la file d'attente, utilisez :
```
!add Never Gonna Give You Up
!resume
```

## 🔧 Configuration du bot

### 🔗 Créer un bot dans le portail des développeurs Discord
Pour ajouter Melodix à un serveur Discord, suivez ces étapes :

1. Créez une application dans le [portail des développeurs Discord](https://discord.com/developers/applications) et obtenez l'`APPLICATION_ID` (dans la section Général).
2. Dans la section Bot, activez les `INTENTIONS DE PRÉSENCE`, `INTENTIONS DE MEMBRES DU SERVEUR` et `INTENTIONS DE CONTENU DES MESSAGES`.
3. Utilisez le lien suivant pour autoriser le bot : `discord.com/oauth2/authorize?client_id=YOUR_APPLICATION_ID&scope=bot&permissions=36727824`
   - Remplacez `YOUR_APPLICATION_ID` par l'ID de votre application de bot obtenu à l'étape 1.
4. Sélectionnez un serveur et cliquez sur "Autoriser".
5. Accordez les autorisations nécessaires à Melodix pour qu'il fonctionne correctement (accès aux canaux de texte et de voix).

Après avoir ajouté le bot, compilez-le à partir des sources ou téléchargez les [binaires compilés](https://github.com/keshon/melodix-player/releases). Les instructions de déploiement Docker sont disponibles dans `docker/README.md`.

### 🛠️ Compilation de Melodix à partir des sources
Ce projet est écrit en Go, assurez-vous donc que votre environnement est prêt. Utilisez les scripts fournis pour compiler Melodix Player à partir des sources :
- `bash-and-run.bat` (ou `.sh` pour Linux) : Compilez la version de débogage et exécutez-la.
- `build-release.bat` (ou `.sh` pour Linux) : Compilez la version de release.
- `assemble-dist.bat` : Compilez la version de release et assemblez-la comme un package de distribution (Windows uniquement).

Renommez `.env.example` en `.env` et stockez votre token de bot Discord dans la variable `DISCORD_BOT_TOKEN`. Installez [FFMPEG](https://ffmpeg.org/) (seules les versions récentes sont prises en charge). Si vous utilisez un FFMPEG portable, spécifiez le chemin dans `DCA_FFMPEG_BINARY_PATH` dans le fichier `.env`.

### 🐳 Déploiement Docker
Pour le déploiement Docker, référez-vous à `docker/README.md` pour des instructions spécifiques.

## 🌐 API REST
Melodix Player fournit plusieurs routes API, susceptibles de changer.

### Routes de guilde
- `GET /guild/ids` : Récupérer les IDs de guilde actives.
- `GET /guild/playing` : Obtenir des informations sur le morceau en cours de lecture dans chaque guilde active.

### Routes d'historique
- `GET /history` : Accéder à l'historique global des morceaux joués.
- `GET /history/:guild_id` : Récupérer l'historique des morceaux joués pour une guilde spécifique.

### Routes d'avatar
- `GET /avatar` : Liste des images disponibles dans le dossier d'avatar.
- `GET /avatar/random` : Récupérer une image aléatoire dans le dossier d'avatar.

### Routes de journal
- `GET /log` : Afficher le journal actuel.
- `GET /log/clear` : Effacer le journal.
- `GET /log/download` : Télécharger le journal sous forme de fichier.

## 🆘 Support
Pour toute question, obtenez de l'aide dans le [serveur Discord officiel](https://discord.gg/NVtdTka8ZT).

## 🏆 Remerciements
Je me suis inspiré de [Muzikas](https://github.com/FabijanZulj/Muzikas), un bot Discord convivial de Fabijan Zulj.

## 📜 Licence
Melodix est sous licence [MIT](https://opensource.org/licenses/MIT).