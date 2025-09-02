import { useState } from "react";
import {
  Button,
  Input,
  VStack,
  Text,
  Link,
  Card,
  Heading,
  Field,
  Alert,
} from "@chakra-ui/react";
import { useSignup } from "@/hooks/auth";
import type { SignupRequest } from "@/lib/api";

interface SignupFormProps {
  onSwitchToSignIn: () => void;
}

export const SignupForm = ({ onSwitchToSignIn }: SignupFormProps) => {
  const [formData, setFormData] = useState<SignupRequest>({
    username: "",
    email: "",
    password: "",
  });

  const signupMutation = useSignup();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    signupMutation.mutate(formData);
  };

  const handleChange = (field: keyof SignupRequest) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: e.target.value,
    }));
  };

  return (
    <Card.Root maxW="md" mx="auto" p="6">
      <Card.Body>
        <VStack gap="4" align="stretch">
          <Heading size="lg" textAlign="center" color="fg">
            Create Account
          </Heading>

          {signupMutation.error && (
            <Alert.Root status="error">
              <Alert.Title>Signup Failed</Alert.Title>
              <Alert.Description>
                {signupMutation.error.message}
              </Alert.Description>
            </Alert.Root>
          )}

          <form onSubmit={handleSubmit}>
            <VStack gap="4" align="stretch">
              <Field.Root required>
                <Field.Label>Username</Field.Label>
                <Input
                  type="text"
                  value={formData.username}
                  onChange={handleChange("username")}
                  placeholder="Enter your username"
                  required
                />
              </Field.Root>

              <Field.Root required>
                <Field.Label>Email</Field.Label>
                <Input
                  type="email"
                  value={formData.email}
                  onChange={handleChange("email")}
                  placeholder="Enter your email"
                  required
                />
              </Field.Root>

              <Field.Root required>
                <Field.Label>Password</Field.Label>
                <Input
                  type="password"
                  value={formData.password}
                  onChange={handleChange("password")}
                  placeholder="Enter your password"
                  required
                />
              </Field.Root>

              <Button
                type="submit"
                colorScheme="blue"
                width="full"
                loading={signupMutation.isPending}
                disabled={signupMutation.isPending}
              >
                {signupMutation.isPending ? "Creating Account..." : "Sign Up"}
              </Button>
            </VStack>
          </form>

          <Text textAlign="center" fontSize="sm" color="fg.muted">
            Already have an account?{" "}
            <Link
              onClick={onSwitchToSignIn}
              color="blue.500"
              cursor="pointer"
              textDecoration="underline"
            >
              Sign in here
            </Link>
          </Text>
        </VStack>
      </Card.Body>
    </Card.Root>
  );
};