function run({ task_id, status }) {
  // Validate status
  const validStatuses = ['todo', 'in_progress', 'done'];
  if (!validStatuses.includes(status)) {
    throw new Error(`Invalid status: ${status}. Must be one of: ${validStatuses.join(', ')}`);
  }
  
  // Get current timestamp
  const now = new Date().toISOString();
  
  // Update task status
  const result = query(`
    UPDATE tasks
    SET status = ?, updated_at = ?
    WHERE id = ? AND deleted_at IS NULL
    RETURNING id
  `, status, now, task_id);
  
  if (!result || result.length === 0) {
    throw new Error(`Task with ID ${task_id} not found or could not be updated`);
  }
  
  return { 
    success: true,
    task_id: result[0][0],
    status
  };
}