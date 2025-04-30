function run({ userId }) {
  let tasksQuery = `
    SELECT 
      t.id, 
      t.title, 
      t.description, 
      t.status, 
      t.priority, 
      t.due_date,
      u.name as user_name
    FROM 
      tasks t
    LEFT JOIN 
      user_tasks ut ON t.id = ut.task_id
    LEFT JOIN 
      users u ON ut.user_id = u.id
    WHERE 
      t.deleted_at IS NULL
  `;

  const params = [];

  if (userId !== null) {
    tasksQuery += ` AND ut.user_id = ?`;
    params.push(userId);
  }

  tasksQuery += ` ORDER BY t.priority DESC, t.due_date ASC`;

  const tasksResult = query(tasksQuery, ...params);

  const tasks = tasksResult.map((row) => ({
    id: row[0],
    title: row[1],
    description: row[2],
    status: row[3],
    priority: row[4],
    dueDate: row[5],
    userName: row[6],
  }));

  return { tasks };
}
