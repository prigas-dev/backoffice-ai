import React, { useState } from "react";
import { Badge, Button, Modal, Table } from "react-bootstrap";

interface Task {
  id: number;
  title: string;
  description: string;
  status: string;
  priority: number;
  due_date: string;
  assigned_user_name: string | null;
  category_name: string | null;
}

interface ComponentProps {
  tasks: Task[];
}

export const Component: React.FC<ComponentProps> = ({ tasks }) => {
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [showModal, setShowModal] = useState(false);

  const handleClose = () => {
    setShowModal(false);
    setSelectedTask(null);
  };

  const handleShow = (task: Task) => {
    setSelectedTask(task);
    setShowModal(true);
  };

  const getPriorityBadge = (priority: number) => {
    switch (priority) {
      case 1:
        return <Badge bg="success">Low</Badge>;
      case 2:
        return <Badge bg="warning">Medium</Badge>;
      case 3:
        return <Badge bg="danger">High</Badge>;
      default:
        return <Badge bg="secondary">Unknown</Badge>;
    }
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "No due date";
    const date = new Date(dateString);
    return date.toLocaleDateString();
  };

  const isOverdue = (dateString: string) => {
    if (!dateString) return false;
    const dueDate = new Date(dateString);
    const today = new Date();
    return dueDate < today;
  };

  return (
    <div className="container mt-4">
      <h2>Todo Tasks</h2>
      {tasks.length === 0 ? (
        <div className="alert alert-info">No todo tasks found.</div>
      ) : (
        <Table striped bordered hover responsive>
          <thead>
            <tr>
              <th>Title</th>
              <th>Priority</th>
              <th>Due Date</th>
              <th>Assigned To</th>
              <th>Category</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {tasks.map((task) => (
              <tr key={task.id}>
                <td>{task.title}</td>
                <td>{getPriorityBadge(task.priority)}</td>
                <td>
                  {isOverdue(task.due_date) ? (
                    <span className="text-danger">
                      {formatDate(task.due_date)} (Overdue)
                    </span>
                  ) : (
                    formatDate(task.due_date)
                  )}
                </td>
                <td>{task.assigned_user_name || "Unassigned"}</td>
                <td>{task.category_name || "Uncategorized"}</td>
                <td>
                  <Button
                    variant="info"
                    size="sm"
                    onClick={() => handleShow(task)}
                  >
                    View Details
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      <Modal show={showModal} onHide={handleClose} size="lg">
        {selectedTask && (
          <>
            <Modal.Header closeButton>
              <Modal.Title>{selectedTask.title}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              <div className="mb-3">
                <h5>Description</h5>
                <p>{selectedTask.description || "No description provided."}</p>
              </div>
              <div className="row mb-3">
                <div className="col-md-4">
                  <h5>Priority</h5>
                  <p>{getPriorityBadge(selectedTask.priority)}</p>
                </div>
                <div className="col-md-4">
                  <h5>Due Date</h5>
                  <p>
                    {isOverdue(selectedTask.due_date) ? (
                      <span className="text-danger">
                        {formatDate(selectedTask.due_date)} (Overdue)
                      </span>
                    ) : (
                      formatDate(selectedTask.due_date)
                    )}
                  </p>
                </div>
                <div className="col-md-4">
                  <h5>Status</h5>
                  <p>
                    <Badge bg="primary">{selectedTask.status}</Badge>
                  </p>
                </div>
              </div>
              <div className="row">
                <div className="col-md-6">
                  <h5>Assigned To</h5>
                  <p>{selectedTask.assigned_user_name || "Unassigned"}</p>
                </div>
                <div className="col-md-6">
                  <h5>Category</h5>
                  <p>{selectedTask.category_name || "Uncategorized"}</p>
                </div>
              </div>
            </Modal.Body>
            <Modal.Footer>
              <Button variant="secondary" onClick={handleClose}>
                Close
              </Button>
            </Modal.Footer>
          </>
        )}
      </Modal>
    </div>
  );
};
