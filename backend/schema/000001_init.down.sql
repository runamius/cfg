DROP INDEX IF EXISTS idx_bookings_active_slot;
DROP INDEX IF EXISTS idx_bookings_user;
DROP INDEX IF EXISTS idx_bookings_slot;
DROP INDEX IF EXISTS idx_slots_room_start;

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
