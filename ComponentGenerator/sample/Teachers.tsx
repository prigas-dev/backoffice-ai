import React, { useState } from "react";
import { Badge, Form, InputGroup, Table } from "react-bootstrap";

interface Teacher {
  id: number;
  first_name: string;
  last_name: string;
  lecture_count: number;
}

interface ComponentProps {
  teachers: Teacher[];
}

export const Component: React.FC<ComponentProps> = ({ teachers }) => {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredTeachers = teachers.filter((teacher) => {
    const fullName = `${teacher.first_name} ${teacher.last_name}`.toLowerCase();
    return fullName.includes(searchTerm.toLowerCase());
  });

  return (
    <div className="p-3">
      <h2>Teachers Who Gave Lectures This Month</h2>

      <InputGroup className="mb-3">
        <InputGroup.Text>Search</InputGroup.Text>
        <Form.Control
          placeholder="Filter by name..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </InputGroup>

      {filteredTeachers.length === 0 ? (
        <p className="text-center mt-4">No teachers found for this month</p>
      ) : (
        <Table striped bordered hover responsive>
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Lecture Count</th>
            </tr>
          </thead>
          <tbody>
            {filteredTeachers.map((teacher) => (
              <tr key={teacher.id}>
                <td>{teacher.id}</td>
                <td>
                  {teacher.first_name} {teacher.last_name}
                </td>
                <td>
                  <Badge bg="primary" pill>
                    {teacher.lecture_count}
                  </Badge>
                </td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <td colSpan={2}>
                <strong>Total Teachers:</strong>
              </td>
              <td>{filteredTeachers.length}</td>
            </tr>
          </tfoot>
        </Table>
      )}
    </div>
  );
};
