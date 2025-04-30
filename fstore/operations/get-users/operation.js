function run() {
  // Get all active users
  const users = query(`
    SELECT id, name, email
    FROM users
    WHERE deleted_at IS NULL
    ORDER BY name ASC
  `);
  
  // Format the users array
  const formattedUsers = users.map(user => ({
    id: user[0],
    name: user[1],
    email: user[2]
  }));
  
  return { users: formattedUsers };
}