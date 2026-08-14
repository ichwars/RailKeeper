CREATE TRIGGER tr_plan_track_reservation_deactivate
AFTER UPDATE OF status ON accessory_reservations
WHEN OLD.status='active' AND NEW.status<>'active'
BEGIN
  UPDATE plan_track_object_reservations
  SET active=0, updated_at=NEW.updated_at
  WHERE reservation_id=NEW.id AND active=1;
END;
