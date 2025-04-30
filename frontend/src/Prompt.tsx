import { useState } from "react";
import { Button, Col, Container, Form, Row, Spinner } from "react-bootstrap";
import { useForm } from "react-hook-form";

type Feature = {
  name: string;
};
export function Prompt() {
  function onFeatureCreated(feature: Feature) {
    location.href = `/feature/${feature.name}`;
  }
  return (
    <Container>
      <CreateViewForm onFeatureCreated={onFeatureCreated} />
    </Container>
  );
}

interface CreateViewFormProps {
  onFeatureCreated(feature: any): void;
}
function CreateViewForm({ onFeatureCreated }: CreateViewFormProps) {
  type CreateViewData = {
    prompt: string;
  };
  const form = useForm({
    defaultValues: {
      prompt: "",
    },
  });

  const [isSubmitting, setIsSubmitting] = useState(false);
  async function onSubmit(data: CreateViewData) {
    if (isSubmitting) {
      return;
    }

    if (!data.prompt) {
      alert("prompt is required");
      return;
    }

    setIsSubmitting(true);

    try {
      const body = new URLSearchParams();
      body.set("prompt", data.prompt);
      const response = await fetch("/create-feature", {
        method: "post",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body,
      });
      if (!response.ok) {
        const message = await response.text();
        throw new Error(`[${response.status}] ${message}`);
      }

      const feature = await response.json();
      console.log(feature);

      onFeatureCreated(feature);
    } catch (error) {
      alert((error as Error).message);
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <Form autoComplete="off" onSubmit={form.handleSubmit(onSubmit)}>
        <Form.Group className="mb-3 mt-3" controlId="prompt">
          <Form.Label>Feature description</Form.Label>
          <Form.Control
            as="textarea"
            rows={5}
            placeholder="ex: Show me all tasks with due date until today"
            {...form.register("prompt")}
          />
        </Form.Group>
        <Button type="submit" variant="primary" disabled={isSubmitting}>
          Submit
        </Button>
      </Form>
      {isSubmitting ? (
        <Loader />
      ) : (
        <Row className="justify-content-center">
          <Col xs="auto">Prompt something to create a new feature</Col>
        </Row>
      )}
    </>
  );
}

function Loader() {
  return (
    <Row className="justify-content-center">
      <Col xs="auto">
        <Spinner animation="grow" variant="success" />
      </Col>
    </Row>
  );
}
