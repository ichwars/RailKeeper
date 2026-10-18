-- Preserve read access of existing planning accounts without adding inventory write access.
INSERT OR IGNORE INTO roles(id,name) VALUES('role-viewer-retirement','Viewer');
INSERT OR IGNORE INTO user_roles(user_id,role_id)
SELECT ur.user_id,(SELECT id FROM roles WHERE name='Viewer') FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE r.name='Planner';
DELETE FROM roles WHERE name='Planner';
