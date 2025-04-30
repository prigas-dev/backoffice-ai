function run({ taskId, status }) {
  // Validate status is one of the allowed values
  const validStatuses = ["todo", "in_progress", "done"];
  if (!validStatuses.includes(status)) {
    throw new Error(
      "Invalid status value. Must be one of: todo, in_progress, done"
    );
  }

  // Check if task exists
  const taskExists = query(
    "SELECT id FROM tasks WHERE id = ? AND deleted_at IS NULL",
    taskId
  );
  if (taskExists.length === 0) {
    throw new Error(`Task with ID ${taskId} not found`);
  }

  // Update the task status
  query(
    'UPDATE tasks SET status = ?, updated_at = datetime("now") WHERE id = ?',
    status,
    taskId
  );

  return { success: true };
}
