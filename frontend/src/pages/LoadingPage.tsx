import { Box, Spinner, Center, Text } from "@chakra-ui/react";

export const LoadingPage = () => {
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
};
