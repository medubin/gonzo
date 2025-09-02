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
import { useSignIn } from "@/hooks/auth";
import type { SignInRequest } from "@/lib/api";

interface SignInFormProps {
  onSwitchToSignUp: () => void;
}

export const SignInForm = ({ onSwitchToSignUp }: SignInFormProps) => {
  const [formData, setFormData] = useState<SignInRequest>({
    email: "",
    password: "",
  });

  const signInMutation = useSignIn();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    signInMutation.mutate(formData);
  };

  const handleChange =
    (field: keyof SignInRequest) =>
    (e: React.ChangeEvent<HTMLInputElement>) => {
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
            Welcome Back
          </Heading>

          {signInMutation.error && (
            <Alert.Root status="error">
              <Alert.Title>Sign In Failed</Alert.Title>
              <Alert.Description>
                {signInMutation.error.message}
              </Alert.Description>
            </Alert.Root>
          )}

          <form onSubmit={handleSubmit}>
            <VStack gap="4" align="stretch">
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
                variant="outline"
                colorPalette="blue"
                width="full"
                loading={signInMutation.isPending}
                disabled={signInMutation.isPending}
              >
                {signInMutation.isPending ? "Signing In..." : "Sign In"}
              </Button>
            </VStack>
          </form>

          <Text textAlign="center" fontSize="sm" color="fg.muted">
            Don't have an account?{" "}
            <Link
              onClick={onSwitchToSignUp}
              color="blue.500"
              cursor="pointer"
              textDecoration="underline"
            >
              Sign up here
            </Link>
          </Text>
        </VStack>
      </Card.Body>
    </Card.Root>
  );
};
