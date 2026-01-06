import { Box, Container, VStack } from "@chakra-ui/react";
import { SignInForm } from "@/components/auth/SignInForm";
import { useNavigate } from "react-router-dom";

export const LoginPage = () => {
  const navigate = useNavigate();

  return (
    <Box minH="100vh" bg="bg" py="12">
      <Container maxW="lg">
        <VStack gap="8" align="stretch">
          <SignInForm onSwitchToSignUp={() => navigate("/signup")} />
        </VStack>
      </Container>
    </Box>
  );
};
