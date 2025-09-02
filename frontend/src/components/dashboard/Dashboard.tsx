import {
  Box,
  Button,
  Container,
  VStack,
  HStack,
  Heading,
  Text,
  Card,
  Badge,
  Separator,
} from "@chakra-ui/react";
import { useAuth, useSignOut } from "@/hooks/auth";
import { ColorModeButton } from "@/components/ui/color-mode";

export const Dashboard = () => {
  const { user } = useAuth(); // Use useAuth instead of useCurrentUser for cleaner logic
  const signOutMutation = useSignOut();

  const handleSignOut = () => {
    signOutMutation.mutate();
  };

  return (
    <Box minH="100vh" bg="bg" py="12">
      <Container maxW="4xl">
        <VStack gap="8" align="stretch">
          {/* Header */}
          <HStack justify="space-between" align="center">
            <Heading size="xl" color="fg">
              Dashboard
            </Heading>
            <HStack gap="2">
              <ColorModeButton />
              <Button
                variant="outline"
                colorScheme="red"
                onClick={handleSignOut}
                loading={signOutMutation.isPending}
                disabled={signOutMutation.isPending}
              >
                {signOutMutation.isPending ? "Signing Out..." : "Sign Out"}
              </Button>
            </HStack>
          </HStack>

          <Separator />

          {/* User Info */}
          {user && (
            <Card.Root>
              <Card.Header>
                <Card.Title>Profile Information</Card.Title>
              </Card.Header>
              <Card.Body>
                <VStack gap="4" align="start">
                  <HStack gap="4" align="center">
                    <Text fontWeight="semibold">Username:</Text>
                    <Text>{user.username}</Text>
                  </HStack>

                  <HStack gap="4" align="center">
                    <Text fontWeight="semibold">Email:</Text>
                    <Text>{user.email}</Text>
                  </HStack>

                  <HStack gap="4" align="center">
                    <Text fontWeight="semibold">Role:</Text>
                    <Badge
                      colorScheme={user.role === "admin" ? "purple" : "blue"}
                    >
                      {user.role || "user"}
                    </Badge>
                  </HStack>

                  {user.createdAt && (
                    <HStack gap="4" align="center">
                      <Text fontWeight="semibold">Member since:</Text>
                      <Text>
                        {new Date(user.createdAt * 1000).toLocaleDateString()}
                      </Text>
                    </HStack>
                  )}
                </VStack>
              </Card.Body>
            </Card.Root>
          )}

          {/* Welcome Message */}
          <Card.Root>
            <Card.Body>
              <VStack gap="4" align="center" textAlign="center">
                <Heading size="md" color="fg">
                  Welcome{user ? `, ${user.username}` : ""}! 🎉
                </Heading>
                <Text color="fg.muted" maxW="2xl">
                  You've successfully signed in to your dashboard. This is a
                  protected page that requires authentication. You can now
                  access all the features available to authenticated users.
                </Text>
              </VStack>
            </Card.Body>
          </Card.Root>

          {/* Quick Actions */}
          <Card.Root>
            <Card.Header>
              <Card.Title>Quick Actions</Card.Title>
            </Card.Header>
            <Card.Body>
              <VStack gap="3" align="stretch">
                <Button variant="outline" width="full">
                  View All Users
                </Button>
                <Button variant="outline" width="full">
                  Update Profile
                </Button>
                <Button variant="outline" width="full">
                  Settings
                </Button>
              </VStack>
            </Card.Body>
          </Card.Root>
        </VStack>
      </Container>
    </Box>
  );
};