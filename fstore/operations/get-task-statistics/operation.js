function run() {
  // Get counts by status
  const statusCounts = query(`
    SELECT status, COUNT(*) as count 
    FROM tasks 
    WHERE deleted_at IS NULL 
    GROUP BY status
  `);
  
  // Get counts by priority
  const priorityCounts = query(`
    SELECT priority, COUNT(*) as count 
    FROM tasks 
    WHERE deleted_at IS NULL 
    GROUP BY priority
    ORDER BY priority
  `);
  
  // Get upcoming tasks (due in the next 7 days)
  const upcomingTasks = query(`
    SELECT id, title, due_date, status
    FROM tasks
    WHERE deleted_at IS NULL
      AND due_date IS NOT NULL
      AND due_date > datetime('now')
      AND due_date <= datetime('now', '+7 days')
    ORDER BY due_date ASC
    LIMIT 5
  `);
  
  // Get total task count
  const totalTasks = query(`
    SELECT COUNT(*) as count
    FROM tasks
    WHERE deleted_at IS NULL
  `);
  
  // Format the results
  const formattedStatusCounts = statusCounts.map(row => ({
    status: row[0],
    count: row[1]
  }));
  
  const formattedPriorityCounts = priorityCounts.map(row => ({
    priority: row[0],
    count: row[1]
  }));
  
  const formattedUpcomingTasks = upcomingTasks.map(row => ({
    id: row[0],
    title: row[1],
    due_date: row[2],
    status: row[3]
  }));
  
  return {
    statusCounts: formattedStatusCounts,
    priorityCounts: formattedPriorityCounts,
    upcomingTasks: formattedUpcomingTasks,
    totalTasks: totalTasks[0][0]
  };
}