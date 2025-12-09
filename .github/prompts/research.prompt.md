---
mode: task-researcher
title: Research options for building a scheudling system for online games
---

I would like to build a scheduling system for online games that allows players to create, join, and manage game sessions. The system should support features such as time zone management, notifications.

There are two classes of users, players and game hosts. Players can browse available game sessions, join sessions, and receive notifications about upcoming games. Game hosts can create and manage game sessions, set player limits, and send updates to participants.

The interface for players will discord via a bot, while game hosts will have access to a web dashboard for managing their sessions.

In a discord server that adds the bot, games will be posted in a cofigurable channel, and players can join or drop from the game by reacting to the message with a specific, pre-created emoji.
The system should also include a notification mechanism to remind players of upcoming games via Discord messages.

When a game is created, the host should be able to specify the game title, description, date and time (with time zone support), maximum number of players, and any special rules or requirements. Once the game is created, it should be listed in a public directory where players can browse and join sessions.

The data about games should be stored in a database that supports efficient querying and retrieval.
