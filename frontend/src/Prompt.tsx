import { useLocation } from "@tanstack/react-router";
import { useState } from "react";
import { Button, Card, Col, Form, Row, Spinner } from "react-bootstrap";
import { useForm } from "react-hook-form";

type Feature = {
  name: string;
};
interface PromptProps {
  promptInstructions?: string;
}
export function Prompt({ promptInstructions }: PromptProps) {
  const currentFeatureName = useCurrentFeatureName();

  function onFeatureCreated(feature: Feature) {
    if (feature.name === currentFeatureName) {
      location.reload();
    } else {
      location.href = `/feature/${feature.name}`;
    }
  }
  return (
    <CreateViewForm
      currentFeatureName={currentFeatureName}
      onFeatureCreated={onFeatureCreated}
      promptInstructions={promptInstructions}
    />
  );
}

export function PromptCornerButton() {
  const [isOpen, setIsOpen] = useState(false);

  const currentFeatureName = useCurrentFeatureName();

  if (currentFeatureName == null) {
    return null;
  }

  return (
    <div
      style={{ zIndex: 1 }}
      className="position-fixed bottom-0 end-0 d-flex flex-column align-items-end m-4"
    >
      {isOpen && (
        <Card className="mb-2 shadow">
          <Card.Body className="">
            <Prompt />
          </Card.Body>
        </Card>
      )}

      <Button
        variant="primary"
        className="rounded-circle d-flex align-items-center justify-content-center"
        style={{ width: "3rem", height: "3rem" }}
        onClick={() => setIsOpen(!isOpen)}
      >
        {isOpen ? (
          <i className="bi-dash-lg"></i>
        ) : (
          <i className="bi-plus-lg"></i>
        )}
      </Button>
    </div>
  );
}

function useCurrentFeatureName(): string | null {
  const location = useLocation();
  const pathParts = location.pathname.split("/").filter(notEmpty);
  if (
    pathParts.length < 2 ||
    pathParts[0] !== "feature" ||
    pathParts[1].length === 0
  ) {
    return null;
  }

  const featureName = pathParts[1];

  return featureName;
}
function notEmpty(value: string): boolean {
  return value.length > 0;
}

interface CreateViewFormProps {
  currentFeatureName: string | null;
  onFeatureCreated(feature: any): void;
  promptInstructions?: string;
}
function CreateViewForm({
  currentFeatureName,
  onFeatureCreated,
  promptInstructions,
}: CreateViewFormProps) {
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
      if (currentFeatureName != null) {
        body.set("feature", currentFeatureName);
      }
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
        promptInstructions != null && (
          <Row className="justify-content-center">
            <Col xs="auto">{promptInstructions}</Col>
          </Row>
        )
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
