import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Form, Spinner } from "react-bootstrap";
import { useForm } from "react-hook-form";

export function Component() {
  const { data: username, isPending, isError, error } = useQueryUsername();

  if (isPending) {
    return (
      <div className="text-center p-4">
        <Spinner animation="border" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="text-center p-4 text-danger">
        Error loading username: {error.message}
      </div>
    );
  }

  return <UsernameForm initialUsername={username} />;
}

type GetUsernameResponseBody = {
  username: string;
};
function useQueryUsername() {
  // Query to fetch the username
  const query = useQuery({
    queryKey: ["username"],
    queryFn: async function (): Promise<string> {
      const responseBody = await executeOperation<{}, GetUsernameResponseBody>(
        "get-username",
        {}
      );
      if (!responseBody.success) {
        throw new Error(responseBody.message);
      }
      return responseBody.result.username;
    },
  });

  return query;
}

function UsernameForm(props: { initialUsername: string }) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<UpdateUsernameRequestBody>({
    defaultValues: {
      username: props.initialUsername,
    },
  });

  const mutation = useMutationUsername();

  function onSubmit(data: UpdateUsernameRequestBody) {
    mutation.mutate(data);
  }

  return (
    <Form onSubmit={handleSubmit(onSubmit)}>
      <Form.Group className="mb-3">
        <Form.Label>Username</Form.Label>
        <Form.Control
          type="text"
          {...register("username", { required: "Username is required" })}
          isInvalid={!!errors.username}
        />
        {errors.username && (
          <Form.Control.Feedback type="invalid">
            {errors.username.message}
          </Form.Control.Feedback>
        )}
      </Form.Group>

      <Button variant="primary" type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? (
          <>
            <Spinner as="span" animation="border" size="sm" className="me-2" />
            Updating...
          </>
        ) : (
          "Update Username"
        )}
      </Button>
    </Form>
  );
}

type UpdateUsernameRequestBody = {
  username: string;
};
type UpdateUsernameResponseBody = {
  username: string;
};
function useMutationUsername() {
  const queryClient = useQueryClient();
  // Mutation to update the username
  const mutation = useMutation({
    mutationFn: async (data: UpdateUsernameRequestBody) => {
      const responseBody = await executeOperation<
        UpdateUsernameRequestBody,
        UpdateUsernameResponseBody
      >("update-username", data);
      if (!responseBody.success) {
        throw new Error(responseBody.message);
      }
      return responseBody.result.username;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["username"] });
    },
  });

  return mutation;
}

type OperationResult<TReturn> =
  | { success: true; result: TReturn }
  | { success: false; message: string };
async function executeOperation<TParameters, TReturn>(
  operationName: string,
  parameters: TParameters
): Promise<OperationResult<TReturn>> {
  const response = await fetch(`/operations/execute/${operationName}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(parameters),
  });

  const operationResult = await response.json();
  return operationResult;
}
