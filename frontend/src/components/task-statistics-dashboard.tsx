import React from 'react';
import { Card, Container, Row, Col, Badge, ListGroup, Spinner, Alert } from 'react-bootstrap';
import { useQuery } from '@tanstack/react-query';

type TaskStatistics = {
  statusCounts: {
    status: string;
    count: number;
  }[];
  priorityCounts: {
    priority: number;
    count: number;
  }[];
  upcomingTasks: {
    id: number;
    title: string;
    due_date: string;
    status: string;
  }[];
  totalTasks: number;
};

type OperationResponse<T> = {
  success: boolean;
  result?: T;
  message?: string;
};

export default function Component() {
  const { data, isLoading, error } = useQuery<TaskStatistics, Error>({
    queryKey: ['taskStatistics'],
    queryFn: async () => {
      const response = await fetch('/operations/execute/get-task-statistics', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ parameters: {} }),
      });
      
      const data = await response.json() as OperationResponse<TaskStatistics>;
      
      if (!data.success || !data.result) {
        throw new Error(data.message || 'Failed to fetch task statistics');
      }
      
      return data.result;
    },
  });

  if (isLoading) {
    return (
      <Container className="mt-4 text-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Loading...</span>
        </Spinner>
      </Container>
    );
  }

  if (error || !data) {
    return (
      <Container className="mt-4">
        <Alert variant="danger">
          Error loading task statistics: {error?.message || 'Unknown error'}
        </Alert>
      </Container>
    );
  }

  const getStatusVariant = (status: string): string => {
    switch (status) {
      case 'done': return 'success';
      case 'in_progress': return 'primary';
      case 'todo': return 'warning';
      default: return 'secondary';
    }
  };

  const getPriorityLabel = (priority: number): string => {
    switch (priority) {
      case 1: return 'Low';
      case 2: return 'Medium';
      case 3: return 'High';
      default: return `Priority ${priority}`;
    }
  };

  const getPriorityVariant = (priority: number): string => {
    switch (priority) {
      case 1: return 'info';
      case 2: return 'warning';
      case 3: return 'danger';
      default: return 'secondary';
    }
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };

  return (
    <Container className="mt-4">
      <h1 className="mb-4">Task Statistics Dashboard</h1>
      
      <Row className="mb-4">
        <Col>
          <Card className="text-center h-100">
            <Card.Body>
              <Card.Title>Total Tasks</Card.Title>
              <h2>{data.totalTasks}</h2>
            </Card.Body>
          </Card>
        </Col>
      </Row>
      
      <Row className="mb-4">
        <Col md={6}>
          <Card className="h-100">
            <Card.Header>Tasks by Status</Card.Header>
            <Card.Body>
              {data.statusCounts.map(item => (
                <div key={item.status} className="d-flex justify-content-between align-items-center mb-2">
                  <div>
                    <Badge bg={getStatusVariant(item.status)} className="me-2">
                      {item.status.replace('_', ' ')}
                    </Badge>
                  </div>
                  <div>
                    <strong>{item.count}</strong> tasks
                  </div>
                </div>
              ))}
            </Card.Body>
          </Card>
        </Col>
        
        <Col md={6}>
          <Card className="h-100">
            <Card.Header>Tasks by Priority</Card.Header>
            <Card.Body>
              {data.priorityCounts.map(item => (
                <div key={item.priority} className="d-flex justify-content-between align-items-center mb-2">
                  <div>
                    <Badge bg={getPriorityVariant(item.priority)} className="me-2">
                      {getPriorityLabel(item.priority)}
                    </Badge>
                  </div>
                  <div>
                    <strong>{item.count}</strong> tasks
                  </div>
                </div>
              ))}
            </Card.Body>
          </Card>
        </Col>
      </Row>
      
      <Row>
        <Col>
          <Card>
            <Card.Header>Upcoming Tasks</Card.Header>
            <ListGroup variant="flush">
              {data.upcomingTasks.length > 0 ? (
                data.upcomingTasks.map(task => (
                  <ListGroup.Item key={task.id} className="d-flex justify-content-between align-items-center">
                    <div>
                      <Badge bg={getStatusVariant(task.status)} className="me-2">
                        {task.status.replace('_', ' ')}
                      </Badge>
                      {task.title}
                    </div>
                    <Badge bg="secondary">
                      Due: {formatDate(task.due_date)}
                    </Badge>
                  </ListGroup.Item>
                ))
              ) : (
                <ListGroup.Item>No upcoming tasks</ListGroup.Item>
              )}
            </ListGroup>
          </Card>
        </Col>
      </Row>
    </Container>
  );
}
