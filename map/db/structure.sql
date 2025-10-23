-- n_db_main.map_chunk_layer definition

CREATE TABLE `map_chunk_layer` (
  `map_id` varchar(64) NOT NULL,
  `layer` varchar(64) NOT NULL,
  `cx` int NOT NULL,
  `cy` int NOT NULL,
  `rev` bigint NOT NULL,
  `payload` longblob NOT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`map_id`,`layer`,`cx`,`cy`),
  KEY `idx_rev` (`map_id`,`layer`,`rev`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- n_db_main.map_entity definition

CREATE TABLE `map_entity` (
  `entity_id` bigint NOT NULL AUTO_INCREMENT,
  `map_id` varchar(64) NOT NULL,
  `type` varchar(32) NOT NULL,
  `x` int NOT NULL,
  `y` int NOT NULL,
  `w` int NOT NULL,
  `h` int NOT NULL,
  `comp_json` json NOT NULL,
  `chunk_x` int NOT NULL,
  `chunk_y` int NOT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`entity_id`),
  KEY `idx_map_type` (`map_id`,`type`),
  KEY `idx_chunk` (`map_id`,`chunk_x`,`chunk_y`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- n_db_main.map_imagelayer definition

CREATE TABLE `map_imagelayer` (
  `map_id` varchar(64) NOT NULL,
  `name` varchar(64) NOT NULL,
  `z_index` int NOT NULL DEFAULT '0',
  `image` varchar(256) NOT NULL,
  `opacity` double NOT NULL DEFAULT '1',
  `repeatx` tinyint(1) NOT NULL DEFAULT '0',
  `repeaty` tinyint(1) NOT NULL DEFAULT '0',
  `x` int NOT NULL DEFAULT '0',
  `y` int NOT NULL DEFAULT '0',
  `visible` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`map_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- n_db_main.map_layer_meta definition

CREATE TABLE `map_layer_meta` (
  `map_id` varchar(64) NOT NULL,
  `layer` varchar(64) NOT NULL,
  `z_index` int NOT NULL DEFAULT '0',
  `kind` enum('base','nature','business','path','deco','other') NOT NULL DEFAULT 'other',
  `style` json DEFAULT NULL,
  PRIMARY KEY (`map_id`,`layer`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


-- n_db_main.map_meta definition

CREATE TABLE `map_meta` (
  `map_id` varchar(64) NOT NULL,
  `header_json` json NOT NULL,
  `tilesets_json` json NOT NULL,
  PRIMARY KEY (`map_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;