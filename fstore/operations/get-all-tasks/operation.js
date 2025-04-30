function run() {
  const result = query(`
    SELECT 
      id, 
      title, 
      description, 
      status, 
      priority, 
      due_date 
    FROM tasks 
    WHERE deleted_at IS NULL
    ORDER BY priority DESC, due_date ASC
  `);

  const tasks = result.map((row) => ({
    id: row[0],
    title: row[1],
    description: row[2],
    status: row[3],
    priority: row[4],
    due_date: row[5],
  }));

  return { tasks };
}
