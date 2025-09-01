import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import type { SignupRequest, SignInRequest, User } from "@/lib/api";
import { toaster } from "@/lib/toaster";

// Auth query keys
export const authKeys = {
  currentUser: () => ["auth", "currentUser"] as const,
  users: () => ["auth", "users"] as const,
};

// Get current user (you'll need to implement this endpoint)
export const useCurrentUser = () => {
  return useQuery({
    queryKey: authKeys.currentUser(),
    queryFn: () => apiClient.getUser({ UserID: 1 }), // You'll need to get the actual user ID
    retry: false, // Don't retry on auth failures
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
};

// Signup mutation
export const useSignup = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: SignupRequest) => apiClient.signup(data),
    onSuccess: (response) => {
      // Set user data in cache
      queryClient.setQueryData(authKeys.currentUser(), { user: response.User });
      
      // Show success message
      toaster.create({
        title: "Account created successfully!",
        description: `Welcome, ${response.User.Username}!`,
        status: "success",
      });
    },
    onError: (error) => {
      toaster.create({
        title: "Signup failed",
        description: error.message,
        status: "error",
      });
    },
  });
};

// Sign in mutation
export const useSignIn = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: SignInRequest) => apiClient.signIn(data),
    onSuccess: (response) => {
      // Set user data in cache
      queryClient.setQueryData(authKeys.currentUser(), { user: response.User });

      // Show success message
      toaster.create({
        title: "Signed in successfully!",
        description: `Welcome back, ${response.User.Username}!`,
        status: "success",
      });
    },
    onError: (error) => {
      toaster.create({
        title: "Sign in failed",
        description: error.message,
        status: "error",
      });
    },
  });
};

// Sign out mutation
export const useSignOut = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.signOut(),
    onSuccess: () => {
      // Clear all auth data
      queryClient.removeQueries({ queryKey: authKeys.currentUser() });
      queryClient.removeQueries({ queryKey: authKeys.users() });

      // Show success message
      toaster.create({
        title: "Signed out successfully",
        description: "See you later!",
        status: "info",
      });
    },
    onError: (error) => {
      toaster.create({
        title: "Sign out failed",
        description: error.message,
        status: "error",
      });
    },
  });
};