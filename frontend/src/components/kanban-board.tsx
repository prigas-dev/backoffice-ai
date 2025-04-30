import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import {
  Badge,
  Button,
  Card,
  Col,
  Container,
  Form,
  Modal,
  Row,
  Spinner,
} from "react-bootstrap";

type Task = {
  id: number;
  title: string;
  description: string;
  status: "todo" | "in_progress" | "done";
  priority: number;
  due_date: string | null;
};

type GetTasksResponse = {
  tasks: Task[];
};

type UpdateTaskStatusRequest = {
  taskId: number;
  status: "todo" | "in_progress" | "done";
};

type UpdateTaskStatusResponse = {
  success: boolean;
};

const statusLabels: Record<string, string> = {
  todo: "To Do",
  in_progress: "In Progress",
  done: "Done",
};

const statusColors: Record<string, string> = {
  todo: "secondary",
  in_progress: "primary",
  done: "success",
};

const priorityLabels: Record<number, string> = {
  1: "Low",
  2: "Medium",
  3: "High",
};

const priorityColors: Record<number, string> = {
  1: "info",
  2: "warning",
  3: "danger",
};

export default function Component() {
  const queryClient = useQueryClient();
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [showModal, setShowModal] = useState(false);

  const { data, isLoading, error } = useQuery<GetTasksResponse>({
    queryKey: ["tasks"],
    queryFn: async () => {
      const response = await fetch("/operations/execute/get-all-tasks", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ parameters: {} }),
      });

      const data = await response.json();
      if (!data.success) {
        throw new Error(data.message);
      }

      return data.result;
    },
  });

  const updateTaskStatusMutation = useMutation<
    UpdateTaskStatusResponse,
    Error,
    UpdateTaskStatusRequest
  >({
    mutationFn: async ({ taskId, status }) => {
      const response = await fetch("/operations/execute/update-task-status", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          parameters: {
            taskId,
            status,
          },
        }),
      });

      const data = await response.json();
      if (!data.success) {
        throw new Error(data.message);
      }

      return data.result;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      setShowModal(false);
    },
  });

  const handleStatusChange = (task: Task) => {
    setSelectedTask(task);
    setShowModal(true);
  };

  const handleUpdateStatus = (status: "todo" | "in_progress" | "done") => {
    if (selectedTask) {
      updateTaskStatusMutation.mutate({
        taskId: selectedTask.id,
        status,
      });
    }
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedTask(null);
  };

  if (isLoading) {
    return (
      <Container className="mt-4 text-center">
        <Spinner animation="border" />
        <p>Loading tasks...</p>
      </Container>
    );
  }

  if (error) {
    return (
      <Container className="mt-4">
        <div className="alert alert-danger">
          Error loading tasks: {error.message}
        </div>
      </Container>
    );
  }

  const todoTasks = data?.tasks.filter((task) => task.status === "todo") || [];
  const inProgressTasks =
    data?.tasks.filter((task) => task.status === "in_progress") || [];
  const doneTasks = data?.tasks.filter((task) => task.status === "done") || [];

  return (
    <Container fluid className="mt-4">
      <h1 className="mb-4">Task Kanban Board</h1>

      <Row>
        <Col md={4}>
          <div className="kanban-column">
            <h4 className="text-center p-2 bg-secondary text-white rounded">
              To Do
            </h4>
            {todoTasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onStatusChange={handleStatusChange}
              />
            ))}
            {todoTasks.length === 0 && (
              <Card className="mb-2 text-center p-3 text-muted">
                <em>No tasks</em>
              </Card>
            )}
          </div>
        </Col>

        <Col md={4}>
          <div className="kanban-column">
            <h4 className="text-center p-2 bg-primary text-white rounded">
              In Progress
            </h4>
            {inProgressTasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onStatusChange={handleStatusChange}
              />
            ))}
            {inProgressTasks.length === 0 && (
              <Card className="mb-2 text-center p-3 text-muted">
                <em>No tasks</em>
              </Card>
            )}
          </div>
        </Col>

        <Col md={4}>
          <div className="kanban-column">
            <h4 className="text-center p-2 bg-success text-white rounded">
              Done
            </h4>
            {doneTasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onStatusChange={handleStatusChange}
              />
            ))}
            {doneTasks.length === 0 && (
              <Card className="mb-2 text-center p-3 text-muted">
                <em>No tasks</em>
              </Card>
            )}
          </div>
        </Col>
      </Row>

      <Modal show={showModal} onHide={handleCloseModal}>
        <Modal.Header closeButton>
          <Modal.Title>Update Task Status</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {selectedTask && (
            <>
              <h5>{selectedTask.title}</h5>
              <p>{selectedTask.description}</p>
              <Form.Group className="mb-3">
                <Form.Label>
                  Current Status:{" "}
                  <Badge bg={statusColors[selectedTask.status]}>
                    {statusLabels[selectedTask.status]}
                  </Badge>
                </Form.Label>
                <div className="d-grid gap-2">
                  <Button
                    variant="secondary"
                    onClick={() => handleUpdateStatus("todo")}
                    disabled={
                      selectedTask.status === "todo" ||
                      updateTaskStatusMutation.isPending
                    }
                  >
                    Move to To Do
                  </Button>
                  <Button
                    variant="primary"
                    onClick={() => handleUpdateStatus("in_progress")}
                    disabled={
                      selectedTask.status === "in_progress" ||
                      updateTaskStatusMutation.isPending
                    }
                  >
                    Move to In Progress
                  </Button>
                  <Button
                    variant="success"
                    onClick={() => handleUpdateStatus("done")}
                    disabled={
                      selectedTask.status === "done" ||
                      updateTaskStatusMutation.isPending
                    }
                  >
                    Move to Done
                  </Button>
                </div>
              </Form.Group>
            </>
          )}
          {updateTaskStatusMutation.isPending && (
            <div className="text-center mt-3">
              <Spinner animation="border" size="sm" /> Updating task status...
            </div>
          )}
          {updateTaskStatusMutation.isError && (
            <div className="alert alert-danger mt-3">
              {updateTaskStatusMutation.error.message}
            </div>
          )}
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={handleCloseModal}>
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>

      <style>{`
        .kanban-column {
          min-height: 300px;
          padding: 10px;
          background-color: #f8f9fa;
          border-radius: 5px;
        }
      `}</style>
    </Container>
  );
}

interface TaskCardProps {
  task: Task;
  onStatusChange: (task: Task) => void;
}

function TaskCard({ task, onStatusChange }: TaskCardProps) {
  const formattedDate = task.due_date
    ? new Date(task.due_date).toLocaleDateString()
    : "No due date";

  return (
    <Card className="mb-2 shadow-sm">
      <Card.Body>
        <Card.Title>{task.title}</Card.Title>
        <Card.Text>{task.description}</Card.Text>
        <div className="d-flex justify-content-between align-items-center">
          <div>
            <Badge bg={statusColors[task.status]} className="me-1">
              {statusLabels[task.status]}
            </Badge>
            <Badge bg={priorityColors[task.priority]} className="me-1">
              {priorityLabels[task.priority]}
            </Badge>
          </div>
          <small className="text-muted">{formattedDate}</small>
        </div>
        <Button
          variant="outline-secondary"
          size="sm"
          className="mt-2 w-100"
          onClick={() => onStatusChange(task)}
        >
          Change Status
        </Button>
      </Card.Body>
    </Card>
  );
}
