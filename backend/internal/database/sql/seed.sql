-- seed.sql

-- Clear existing data
TRUNCATE TABLE sync_events, playlist_items, media, windows RESTART IDENTITY CASCADE;

-- Seed Media Items
INSERT INTO media (id, name, type, url, duration) VALUES
(1, 'Cyberpunk City Skyline', 'image', 'https://images.unsplash.com/photo-1519501025264-65ba15a82390?auto=format&fit=crop&w=1200&q=80', 10),
(2, 'Aurora Borealis Lights', 'image', 'https://images.unsplash.com/photo-1531366936337-7c912a4589a7?auto=format&fit=crop&w=1200&q=80', 15),
(3, 'Ocean Waves Looping', 'video', 'https://media.w3.org/2010/05/video/movie_300.mp4', 20),
(4, 'Neon Abstract Animation', 'video', 'https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4', 15),
(5, 'Mountain Peak Sunset', 'image', 'https://images.unsplash.com/photo-1464822759023-fed622ff2c3b?auto=format&fit=crop&w=1200&q=80', 10),
(6, 'Intermission Screen', 'blank', 'blank://black', 5),
(7, 'Cosmic Big Buck Sequence', 'video', 'https://www.w3schools.com/html/mov_bbb.mp4', 25),
(8, 'Forest Mist Morning', 'image', 'https://images.unsplash.com/photo-1448375240586-882707db888b?auto=format&fit=crop&w=1200&q=80', 12),
(9, 'Elephants Dream Loop', 'video', 'https://www.w3schools.com/html/movie.mp4', 18),
(10, 'Tears of Steel Clip', 'video', 'https://media.w3.org/2010/05/sintel/trailer.mp4', 15),
(11, 'Desert Dune Sunrise', 'image', 'https://images.unsplash.com/photo-1509316975850-ff9c5deb0cd9?auto=format&fit=crop&w=1200&q=80', 10),
(12, 'Sintel Animated Teaser', 'video', 'https://media.w3.org/2010/05/video/movie_300.mp4', 20),
(13, 'Global Broadcast Synchronization', 'video', 'https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4', 30),
(14, 'Breaking Live Flash Announcement', 'image', 'https://images.unsplash.com/photo-1585829365295-ab7cd400c167?auto=format&fit=crop&w=1200&q=80', 20);

SELECT setval('media_id_seq', (SELECT MAX(id) FROM media));

-- Seed Windows
INSERT INTO windows (id, name) VALUES
(1, 'Window 1 - Main Stage Left'),
(2, 'Window 2 - Main Stage Right'),
(3, 'Window 3 - Concourse Display'),
(4, 'Window 4 - VIP Lounge');

SELECT setval('windows_id_seq', (SELECT MAX(id) FROM windows));

-- Seed Playlist Items
-- Window 1: M1 -> M2 -> M3
INSERT INTO playlist_items (window_id, media_id, position) VALUES
(1, 1, 0),
(1, 2, 1),
(1, 3, 2);

-- Window 2: M4 -> M5 -> M6
INSERT INTO playlist_items (window_id, media_id, position) VALUES
(2, 4, 0),
(2, 5, 1),
(2, 6, 2);

-- Window 3: M7 -> M8 -> M9
INSERT INTO playlist_items (window_id, media_id, position) VALUES
(3, 7, 0),
(3, 8, 1),
(3, 9, 2);

-- Window 4: M10 -> M11 -> M12
INSERT INTO playlist_items (window_id, media_id, position) VALUES
(4, 10, 0),
(4, 11, 1),
(4, 12, 2);

SELECT setval('playlist_items_id_seq', (SELECT MAX(id) FROM playlist_items));
