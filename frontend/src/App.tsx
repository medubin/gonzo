import { Box, Spinner, Center, Text } from "@chakra-ui/react";
import { useAuth } from "@/hooks/auth";
import { AuthPage } from "@/components/auth/AuthPage";
import { Dashboard } from "@/components/dashboard/Dashboard";
import { Toaster } from "@/components/ui/toaster";

function App() {
  const { isAuthenticated, isLoading } = useAuth();

  // Show loading spinner while checking auth status
  if (isLoading) {
    return (
      <Box minH="100vh" bg="bg">
        <Center h="100vh">
          <Box textAlign="center">
            <Spinner size="xl" color="blue.500" mb="4" />
            <Text color="fg.muted">Loading...</Text>
          </Box>
        </Center>
      </Box>
    );
  }

  return (
    <Box minH="100vh" bg="bg">
      {isAuthenticated ? (
        <Dashboard />
      ) : (
        <AuthPage />
      )}
      <Toaster />
    </Box>
  );
}

export default App;
