CREATE TABLE IF NOT EXISTS `users` (
    `id` text,
    `voice` text,
    `speed` real,
    `tone` real,
    `intone` real,
    `threshold` real,
    `all_pass` real,
    `volume` real,
    `created_at` datetime,
    `updated_at` datetime,
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `bots` (
    `id` text,
    `guild_id` text,
    `wav` text,
    `created_at` datetime,
    `updated_at` datetime,
    PRIMARY KEY (`id`,`guild_id`),
    CONSTRAINT `fk_guilds_bots` FOREIGN KEY (`guild_id`) REFERENCES `guilds`(`id`)
);

CREATE TABLE IF NOT EXISTS `words` (
    `guild_id` text,
    `before` text,
    `after` text,
    `created_at` datetime,
    `updated_at` datetime,
    PRIMARY KEY (`guild_id`,`before`),
    CONSTRAINT `fk_guilds_words` FOREIGN KEY (`guild_id`) REFERENCES `guilds`(`id`)
);

CREATE TABLE IF NOT EXISTS `guilds` (
    `id` text,
    `created_at` datetime,
    `updated_at` datetime,
    PRIMARY KEY (`id`)
);
