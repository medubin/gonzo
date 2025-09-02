import { useState, useEffect } from "react";
import { Box, Container, VStack } from "@chakra-ui/react";
import { SignupForm } from "./SignupForm";
import { SignInForm } from "./SignInForm";

export const AuthPage = () => {
  const [isSignUp, setIsSignUp] = useState(false);

  // Reset to sign-in form when component mounts (e.g., after sign out)
  useEffect(() => {
    setIsSignUp(false);
  }, []);

  return (
    <Box minH="100vh" bg="bg" py="12">
      <Container maxW="lg">
        <VStack gap="8" align="stretch">
          {isSignUp ? (
            <SignupForm
              onSwitchToSignIn={() => setIsSignUp(false)}
            />
          ) : (
            <SignInForm
              onSwitchToSignUp={() => setIsSignUp(true)}
            />
          )}
        </VStack>
      </Container>
    </Box>
  );
};