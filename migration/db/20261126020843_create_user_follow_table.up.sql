CREATE TABLE `user_follow` (
  `follower_id` INT NOT NULL,
  `following_id` INT NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`follower_id`, `following_id`),
  KEY `idx_user_follow_following_id` (`following_id`),
  CONSTRAINT `fk_user_follow_follower`
    FOREIGN KEY (`follower_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_follow_following`
    FOREIGN KEY (`following_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
);