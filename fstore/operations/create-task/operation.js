function run({ title, description, priority, due_date, user_ids }) {
  // Get current timestamp
  const now = new Date().toISOString();
  
  // Insert new task
  const taskResult = query(`
    INSERT INTO tasks (title, description, status, priority, due_date, created_at, updated_at)
    VALUES (?, ?, 'todo', ?, ?, ?, ?)
    RETURNING id
  `, title, description, priority, due_date || null, now, now);
  
  if (!taskResult || taskResult.length === 0) {
    throw new Error('Failed to create task');
  }
  
  const taskId = taskResult[0][0];
  
  // Assign users to the task if provided
  if (user_ids && user_ids.length > 0) {
    for (const userId of user_ids) {
      try {
        query(`
          INSERT INTO user_tasks (user_id, task_id)
          VALUES (?, ?)
        `, userId, taskId);
      } catch (error) {
        // Log error but continue with other users
        console.error(`Failed to assign user ${userId} to task ${taskId}:`, error);
      }
    }
  }
  
  return { 
    success: true,
    task_id: taskId
  };
}