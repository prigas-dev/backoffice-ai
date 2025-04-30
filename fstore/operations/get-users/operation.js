function run() {
  const usersResult = query(`
    SELECT id, name, email 
    FROM users 
    WHERE deleted_at IS NULL 
    ORDER BY name ASC
  `);

  const users = usersResult.map((row) => ({
    id: row[0],
    name: row[1],
    email: row[2],
  }));

  return { users };
}
