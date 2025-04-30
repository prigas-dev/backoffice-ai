import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Table, Form, Spinner, Alert, Badge, Container, Row, Col, Card } from 'react-bootstrap';

type User = {
  id: number;
  name: string;
  email: string;
};

type Task = {
  id: number;
  title: string;
  description: string;
  status: 'done' | 'todo' | 'in_progress';
  priority: number;
  dueDate: string | null;
  userName: string | null;
};

type GetUsersResponse = {
  users: User[];
};

type GetTasksByUserIdParams = {
  userId: number | null;
};

type GetTasksByUserIdResponse = {
  tasks: Task[];
};

async function fetchUsers(): Promise<GetUsersResponse> {
  const response = await fetch('/operations/execute/get-users', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ parameters: {} }),
  });
  
  const data = await response.json();
  if (!data.success) {
    throw new Error(data.message);
  }
  
  return data.result;
}

async function fetchTasksByUserId(userId: number | null): Promise<GetTasksByUserIdResponse> {
  const response = await fetch('/operations/execute/get-tasks-by-user-id', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ parameters: { userId } }),
  });
  
  const data = await response.json();
  if (!data.success) {
    throw new Error(data.message);
  }
  
  return data.result;
}

function getStatusBadgeVariant(status: string): string {
  switch (status) {
    case 'done':
      return 'success';
    case 'in_progress':
      return 'warning';
    case 'todo':
      return 'secondary';
    default:
      return 'light';
  }
}

function formatDate(dateString: string | null): string {
  if (!dateString) return 'No due date';
  return new Date(dateString).toLocaleDateString();
}

function getPriorityLabel(priority: number): string {
  switch (priority) {
    case 1:
      return 'Low';
    case 2:
      return 'Medium';
    case 3:
      return 'High';
    default:
      return `${priority}`;
  }
}

export default function Component() {
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  
  const { data: usersData, isLoading: isLoadingUsers, error: usersError } = useQuery({
    queryKey: ['users'],
    queryFn: fetchUsers
  });
  
  const { data: tasksData, isLoading: isLoadingTasks, error: tasksError } = useQuery({
    queryKey: ['tasks', selectedUserId],
    queryFn: () => fetchTasksByUserId(selectedUserId)
  });
  
  const handleUserChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    setSelectedUserId(value === '' ? null : parseInt(value, 10));
  };
  
  if (isLoadingUsers) {
    return (
      <div className="d-flex justify-content-center p-5">
        <Spinner animation="border" />
      </div>
    );
  }
  
  if (usersError) {
    return <Alert variant="danger">Error loading users: {(usersError as Error).message}</Alert>;
  }
  
  return (
    <Container fluid className="p-3">
      <Card className="mb-4">
        <Card.Body>
          <Card.Title>Task List</Card.Title>
          <Row className="mb-3">
            <Col md={6}>
              <Form.Group>
                <Form.Label>Filter by User</Form.Label>
                <Form.Select 
                  value={selectedUserId === null ? '' : selectedUserId.toString()} 
                  onChange={handleUserChange}
                >
                  <option value="">All Users</option>
                  {usersData?.users.map(user => (
                    <option key={user.id} value={user.id.toString()}>
                      {user.name}
                    </option>
                  ))}
                </Form.Select>
              </Form.Group>
            </Col>
          </Row>
          
          {isLoadingTasks ? (
            <div className="d-flex justify-content-center p-3">
              <Spinner animation="border" />
            </div>
          ) : tasksError ? (
            <Alert variant="danger">Error loading tasks: {(tasksError as Error).message}</Alert>
          ) : tasksData?.tasks.length === 0 ? (
            <Alert variant="info">No tasks found for the selected criteria.</Alert>
          ) : (
            <Table responsive striped hover>
              <thead>
                <tr>
                  <th>Title</th>
                  <th>Description</th>
                  <th>Status</th>
                  <th>Priority</th>
                  <th>Due Date</th>
                  <th>Assigned To</th>
                </tr>
              </thead>
              <tbody>
                {tasksData?.tasks.map(task => (
                  <tr key={task.id}>
                    <td>{task.title}</td>
                    <td>{task.description}</td>
                    <td>
                      <Badge bg={getStatusBadgeVariant(task.status)}>
                        {task.status.replace('_', ' ')}
                      </Badge>
                    </td>
                    <td>{getPriorityLabel(task.priority)}</td>
                    <td>{formatDate(task.dueDate)}</td>
                    <td>{task.userName || 'Unassigned'}</td>
                  </tr>
                ))}
              </tbody>
            </Table>
          )}
        </Card.Body>
      </Card>
    </Container>
  );
}
