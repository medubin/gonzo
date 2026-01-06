import { Box, Container, VStack } from "@chakra-ui/react";
import { SignupForm } from "@/components/auth/SignupForm";
import { useNavigate } from "react-router-dom";

export const SignupPage = () => {
  const navigate = useNavigate();

  return (
    <Box minH="100vh" bg="bg" py="12">
      <Container maxW="lg">
        <VStack gap="8" align="stretch">
          <SignupForm onSwitchToSignIn={() => navigate("/login")} />
        </VStack>
      </Container>
    </Box>
  );
};
