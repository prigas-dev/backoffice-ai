import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createRootRoute,
  createRoute,
  createRouter,
  Link,
  Outlet,
  RouterProvider,
  useLocation,
} from "@tanstack/react-router";
import { StrictMode, Suspense } from "react";
import { Col, Container, Nav, Navbar, Row, Spinner } from "react-bootstrap";
import { createRoot } from "react-dom/client";
import { Prompt, PromptCornerButton } from "./Prompt";
import { FeatureComponents } from "./features";

window.addEventListener("DOMContentLoaded", async function () {
  const features = await getFeatures();

  const rootRoute = createRootRoute({
    component: () => <RootComponent features={features} />,
  });

  const featureRoutes = features.map((feature) => {
    const FeatureComponent = FeatureComponents[feature.name];
    return createRoute({
      getParentRoute: () => rootRoute,
      path: `/feature/${feature.name}`,
      component: () => (
        <Suspense fallback={<Loader />}>
          <FeatureComponent />
        </Suspense>
      ),
    });
  });

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: () => (
      <Container>
        <Prompt promptInstructions="Prompt something to create a new feature" />
      </Container>
    ),
  });

  const routeTree = rootRoute.addChildren([indexRoute, ...featureRoutes]);

  const router = createRouter({ routeTree });

  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: {
        retry: false,
      },
      queries: {
        retry: false,
      },
    },
  });

  const rootElement = document.getElementById("root");
  if (rootElement == null) {
    alert("could not find #root");
    throw new Error("could not find #root");
  }
  const root = createRoot(rootElement);
  root.render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>
  );
});

function RootComponent({ features }: { features: FeatureManifest[] }) {
  const location = useLocation();
  const pathname = location.pathname;

  return (
    <>
      <Navbar bg="primary" data-bs-theme="dark">
        <Container>
          <Navbar.Brand href="#home">Backoffice AI</Navbar.Brand>
          <Nav className="me-auto">
            <Nav.Link as={Link} to="/">
              Home
            </Nav.Link>
            {features.map((feature) => (
              <Nav.Link
                key={feature.name}
                as={Link}
                to={`/feature/${feature.name}`}
              >
                {feature.label}
              </Nav.Link>
            ))}
          </Nav>
        </Container>
      </Navbar>

      <PromptCornerButton />

      <Outlet />
    </>
  );
}

type FeatureManifest = {
  name: string;
  label: string;
  description: string;
  operations: string[];
};
async function getFeatures() {
  const response = await fetch("/get-all-features");
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message);
  }

  const body = await response.json();
  const features: FeatureManifest[] = body.features;

  return features;
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
