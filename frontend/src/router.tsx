import { createBrowserRouter, redirect, Navigate, Outlet } from "react-router-dom";
import { LoginPage } from "@/pages/LoginPage";
import { SignupPage } from "@/pages/SignupPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { queryClient } from "@/lib/queryClient";
import { authKeys } from "@/hooks/auth";
import type { GetUserResponse } from "@/lib/api";
import { apiClient } from "@/lib/api";
import { Box } from "@chakra-ui/react";

// Root layout component
const RootLayout = () => {
  return (
    <Box minH="100vh" bg="bg">
      <Outlet />
    </Box>
  );
};

// Loader for protected routes - checks authentication
const protectedLoader = async () => {
  // Check if we have cached auth data
  const cachedData = queryClient.getQueryData<GetUserResponse>(authKeys.currentUser());

  if (cachedData?.user) {
    return cachedData;
  }

  // Try to restore session from cookie
  try {
    const userData = await apiClient.getCurrentUser();
    // Cache the user data
    queryClient.setQueryData(authKeys.currentUser(), userData);
    return userData;
  } catch (error) {
    // Not authenticated, redirect to login
    return redirect("/login");
  }
};

// Loader for auth pages (login/signup) - redirect if already authenticated
const authLoader = async () => {
  // Check if we have cached auth data
  const cachedData = queryClient.getQueryData<GetUserResponse>(authKeys.currentUser());

  if (cachedData?.user) {
    // Already authenticated, redirect to dashboard
    return redirect("/dashboard");
  }

  // Try to restore session from cookie
  try {
    const userData = await apiClient.getCurrentUser();
    // Cache the user data
    queryClient.setQueryData(authKeys.currentUser(), userData);
    // Already authenticated, redirect to dashboard
    return redirect("/dashboard");
  } catch (error) {
    // Not authenticated, allow access to auth pages
    return null;
  }
};

export const router = createBrowserRouter([
  {
    path: "/",
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/login" replace />,
      },
      {
        path: "login",
        element: <LoginPage />,
        loader: authLoader,
      },
      {
        path: "signup",
        element: <SignupPage />,
        loader: authLoader,
      },
      {
        path: "dashboard",
        element: <DashboardPage />,
        loader: protectedLoader,
      },
    ],
  },
]);
