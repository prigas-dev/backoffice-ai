function run() {
  // Get all tasks with their details
  const tasks = query(`
    SELECT 
      t.id, 
      t.title, 
      t.description, 
      t.status, 
      t.priority, 
      t.due_date
    FROM tasks t
    WHERE t.deleted_at IS NULL
    ORDER BY t.priority ASC, t.due_date ASC
  `);
  
  // Create an array to hold the full task objects
  const tasksWithUsers = [];
  
  // For each task, get the assigned users
  for (const task of tasks) {
    const [id, title, description, status, priority, due_date] = task;
    
    // Get users assigned to this task
    const taskUsers = query(`
      SELECT u.id, u.name, u.email
      FROM users u
      JOIN user_tasks ut ON u.id = ut.user_id
      WHERE ut.task_id = ? AND u.deleted_at IS NULL
    `, id);
    
    // Format users array
    const users = taskUsers.map(user => ({
      id: user[0],
      name: user[1],
      email: user[2]
    }));
    
    // Add the task with its users to the result array
    tasksWithUsers.push({
      id,
      title,
      description,
      status,
      priority,
      due_date,
      users
    });
  }
  
  return { tasks: tasksWithUsers };
}