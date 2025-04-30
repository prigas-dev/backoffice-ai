import React, { useState } from 'react';
import { Card, Button, Container, Row, Col, Form, Badge, Modal, Spinner, Dropdown } from 'react-bootstrap';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';

// Types
type Task = {
  id: number;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: number;
  due_date: string | null;
  users: User[];
};

type User = {
  id: number;
  name: string;
  email: string;
};

type Category = {
  id: number;
  name: string;
};

// Main Component
export default function Component() {
  const [showNewTaskModal, setShowNewTaskModal] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string | null>(null);
  const [userFilter, setUserFilter] = useState<number | null>(null);
  
  const { data: tasks, isLoading: tasksLoading } = useGetTasks();
  const { data: users, isLoading: usersLoading } = useGetUsers();
  
  const handleCloseNewTaskModal = () => setShowNewTaskModal(false);
  const handleShowNewTaskModal = () => setShowNewTaskModal(true);
  
  // Filter tasks based on selected filters
  const filteredTasks = tasks?.filter(task => {
    // Apply status filter if set
    if (statusFilter && task.status !== statusFilter) {
      return false;
    }
    
    // Apply user filter if set
    if (userFilter && !task.users.some(user => user.id === userFilter)) {
      return false;
    }
    
    return true;
  });
  
  // Group tasks by status
  const todoTasks = filteredTasks?.filter(task => task.status === 'todo') || [];
  const inProgressTasks = filteredTasks?.filter(task => task.status === 'in_progress') || [];
  const doneTasks = filteredTasks?.filter(task => task.status === 'done') || [];
  
  if (tasksLoading || usersLoading) {
    return (
      <div className="d-flex justify-content-center align-items-center" style={{ height: '100vh' }}>
        <Spinner animation="border" />
      </div>
    );
  }
  
  return (
    <Container fluid className="p-4">
      <Row className="mb-4">
        <Col>
          <h1>Kanban Board</h1>
        </Col>
        <Col xs="auto" className="d-flex align-items-center">
          <Button variant="primary" onClick={handleShowNewTaskModal} className="me-2">
            New Task
          </Button>
          
          <Form.Group className="me-2" style={{ width: '200px' }}>
            <Form.Select 
              value={statusFilter || ''} 
              onChange={(e) => setStatusFilter(e.target.value || null)}
            >
              <option value="">All Statuses</option>
              <option value="todo">To Do</option>
              <option value="in_progress">In Progress</option>
              <option value="done">Done</option>
            </Form.Select>
          </Form.Group>
          
          <Form.Group style={{ width: '200px' }}>
            <Form.Select 
              value={userFilter || ''} 
              onChange={(e) => setUserFilter(e.target.value ? Number(e.target.value) : null)}
            >
              <option value="">All Users</option>
              {users?.map(user => (
                <option key={user.id} value={user.id}>{user.name}</option>
              ))}
            </Form.Select>
          </Form.Group>
        </Col>
      </Row>
      
      <Row>
        <Col md={4}>
          <KanbanColumn 
            title="To Do" 
            tasks={todoTasks} 
            users={users || []} 
            status="todo" 
          />
        </Col>
        <Col md={4}>
          <KanbanColumn 
            title="In Progress" 
            tasks={inProgressTasks} 
            users={users || []} 
            status="in_progress" 
          />
        </Col>
        <Col md={4}>
          <KanbanColumn 
            title="Done" 
            tasks={doneTasks} 
            users={users || []} 
            status="done" 
          />
        </Col>
      </Row>
      
      {/* New Task Modal */}
      <Modal show={showNewTaskModal} onHide={handleCloseNewTaskModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>Create New Task</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <TaskForm users={users || []} onSuccess={handleCloseNewTaskModal} />
        </Modal.Body>
      </Modal>
    </Container>
  );
}

// Kanban Column Component
function KanbanColumn({ 
  title, 
  tasks, 
  users,
  status 
}: { 
  title: string; 
  tasks: Task[]; 
  users: User[];
  status: 'todo' | 'in_progress' | 'done';
}) {
  return (
    <div className="kanban-column">
      <h4 className="mb-3">{title} ({tasks.length})</h4>
      <div 
        className="p-2" 
        style={{ 
          backgroundColor: '#f5f5f5', 
          borderRadius: '5px',
          minHeight: '500px'
        }}
      >
        {tasks.map(task => (
          <TaskCard 
            key={task.id} 
            task={task} 
            users={users} 
          />
        ))}
      </div>
    </div>
  );
}

// Task Card Component
function TaskCard({ task, users }: { task: Task; users: User[] }) {
  const [showTaskModal, setShowTaskModal] = useState(false);
  const { mutate: updateTaskStatus } = useUpdateTaskStatus();
  
  const handleCloseTaskModal = () => setShowTaskModal(false);
  const handleShowTaskModal = () => setShowTaskModal(true);
  
  const handleMoveTask = (newStatus: 'todo' | 'in_progress' | 'done') => {
    updateTaskStatus({ taskId: task.id, status: newStatus });
  };
  
  // Format due date if exists
  const formattedDueDate = task.due_date ? new Date(task.due_date).toLocaleDateString() : 'No due date';
  
  // Priority badge color
  const getPriorityBadge = (priority: number) => {
    switch(priority) {
      case 1: return 'danger';
      case 2: return 'warning';
      case 3: return 'info';
      default: return 'secondary';
    }
  };
  
  return (
    <>
      <Card className="mb-2 task-card" onClick={handleShowTaskModal}>
        <Card.Body>
          <Card.Title>{task.title}</Card.Title>
          <Card.Text className="text-truncate">{task.description}</Card.Text>
          <div className="d-flex justify-content-between align-items-center">
            <Badge bg={getPriorityBadge(task.priority)}>
              Priority: {task.priority}
            </Badge>
            <small className="text-muted">{formattedDueDate}</small>
          </div>
          {task.users.length > 0 && (
            <div className="mt-2">
              <small>Assigned to: {task.users.map(user => user.name).join(', ')}</small>
            </div>
          )}
        </Card.Body>
      </Card>
      
      {/* Task Detail Modal */}
      <Modal show={showTaskModal} onHide={handleCloseTaskModal}>
        <Modal.Header closeButton>
          <Modal.Title>{task.title}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p><strong>Description:</strong> {task.description}</p>
          <p><strong>Status:</strong> {task.status.replace('_', ' ')}</p>
          <p><strong>Priority:</strong> {task.priority}</p>
          <p><strong>Due Date:</strong> {formattedDueDate}</p>
          <p>
            <strong>Assigned to:</strong> {task.users.length > 0 
              ? task.users.map(user => user.name).join(', ')
              : 'Unassigned'}
          </p>
        </Modal.Body>
        <Modal.Footer>
          <Dropdown>
            <Dropdown.Toggle variant="primary" id="dropdown-move">
              Move Task
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item 
                onClick={() => handleMoveTask('todo')} 
                disabled={task.status === 'todo'}
              >
                To Do
              </Dropdown.Item>
              <Dropdown.Item 
                onClick={() => handleMoveTask('in_progress')} 
                disabled={task.status === 'in_progress'}
              >
                In Progress
              </Dropdown.Item>
              <Dropdown.Item 
                onClick={() => handleMoveTask('done')} 
                disabled={task.status === 'done'}
              >
                Done
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="secondary" onClick={handleCloseTaskModal}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
}

// Task Form Component
function TaskForm({ users, onSuccess }: { users: User[], onSuccess: () => void }) {
  const { register, handleSubmit, formState: { errors }, reset } = useForm<{
    title: string;
    description: string;
    priority: number;
    due_date: string;
    user_ids: number[];
  }>();
  
  const { mutate: createTask, isPending } = useCreateTask();
  
  const onSubmit = (data: any) => {
    // Convert user_ids to array if it's a single value
    const user_ids = Array.isArray(data.user_ids) 
      ? data.user_ids.map(Number) 
      : data.user_ids 
        ? [Number(data.user_ids)] 
        : [];
    
    createTask({
      ...data,
      priority: Number(data.priority),
      user_ids
    }, {
      onSuccess: () => {
        reset();
        onSuccess();
      }
    });
  };
  
  return (
    <Form onSubmit={handleSubmit(onSubmit)}>
      <Form.Group className="mb-3">
        <Form.Label>Title</Form.Label>
        <Form.Control 
          type="text" 
          {...register('title', { required: 'Title is required' })} 
          isInvalid={!!errors.title}
        />
        {errors.title && (
          <Form.Control.Feedback type="invalid">
            {errors.title.message}
          </Form.Control.Feedback>
        )}
      </Form.Group>
      
      <Form.Group className="mb-3">
        <Form.Label>Description</Form.Label>
        <Form.Control 
          as="textarea" 
          rows={3} 
          {...register('description')} 
        />
      </Form.Group>
      
      <Form.Group className="mb-3">
        <Form.Label>Priority (1-High, 2-Medium, 3-Low)</Form.Label>
        <Form.Select 
          {...register('priority', { required: 'Priority is required' })} 
          isInvalid={!!errors.priority}
        >
          <option value="1">High</option>
          <option value="2">Medium</option>
          <option value="3">Low</option>
        </Form.Select>
        {errors.priority && (
          <Form.Control.Feedback type="invalid">
            {errors.priority.message}
          </Form.Control.Feedback>
        )}
      </Form.Group>
      
      <Form.Group className="mb-3">
        <Form.Label>Due Date</Form.Label>
        <Form.Control 
          type="date" 
          {...register('due_date')} 
        />
      </Form.Group>
      
      <Form.Group className="mb-3">
        <Form.Label>Assign Users</Form.Label>
        <Form.Select 
          multiple 
          {...register('user_ids')} 
        >
          {users.map(user => (
            <option key={user.id} value={user.id}>{user.name}</option>
          ))}
        </Form.Select>
      </Form.Group>
      
      <Button variant="primary" type="submit" disabled={isPending}>
        {isPending ? (
          <>
            <Spinner as="span" animation="border" size="sm" className="me-2" />
            Creating...
          </>
        ) : 'Create Task'}
      </Button>
    </Form>
  );
}

// API Hooks
function useGetTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      const response = await fetch('/operations/execute/get-tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ parameters: {} }),
      });
      
      const data = await response.json();
      if (!data.success) {
        throw new Error(data.message);
      }
      
      return data.result.tasks;
    },
  });
}

function useGetUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await fetch('/operations/execute/get-users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ parameters: {} }),
      });
      
      const data = await response.json();
      if (!data.success) {
        throw new Error(data.message);
      }
      
      return data.result.users;
    },
  });
}

function useCreateTask() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (taskData: {
      title: string;
      description: string;
      priority: number;
      due_date?: string;
      user_ids: number[];
    }) => {
      const response = await fetch('/operations/execute/create-task', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ parameters: taskData }),
      });
      
      const data = await response.json();
      if (!data.success) {
        throw new Error(data.message);
      }
      
      return data.result;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}

function useUpdateTaskStatus() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ taskId, status }: { taskId: number; status: string }) => {
      const response = await fetch('/operations/execute/update-task-status', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          parameters: { 
            task_id: taskId, 
            status 
          } 
        }),
      });
      
      const data = await response.json();
      if (!data.success) {
        throw new Error(data.message);
      }
      
      return data.result;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}
