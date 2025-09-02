import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import type { SignupRequest, SignInRequest, GetUserResponse } from "@/lib/api";
import { toaster } from "@/lib/toaster";
import { queryClient } from "@/lib/queryClient";

// Auth query keys
export const authKeys = {
  currentUser: () => ["auth", "currentUser"] as const,
  users: () => ["auth", "users"] as const,
};

// Hook to check if user is authenticated and get current user data
export const useCurrentUser = () => {
  return useQuery<GetUserResponse>({
    queryKey: authKeys.currentUser(),
    queryFn: async () => {
      // First check if we have cached data
      const cachedData = queryClient.getQueryData<GetUserResponse>(authKeys.currentUser());
      if (cachedData) {
        return cachedData;
      }

      // If no cached data, try to restore session from cookie by making an API call
      // This will work if the session cookie is still valid
      try {
        const userData = await apiClient.getCurrentUser();
        return userData;
      } catch (error) {
        throw new Error(`No user session found ${error}`);
      }
    },
    retry: false, // Don't retry on auth failures
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 10, // Keep in cache longer for session restoration
  });
};

// Helper hook to check authentication status
export const useAuth = () => {
  const { data, isLoading, error } = useCurrentUser();
  
  // Treat "No user session found" as normal unauthenticated state, not an error
  const isNoSessionError = error?.message === "No user session found";
  const actualError = isNoSessionError ? null : error;
  
  // Check if data is explicitly null (set during signout) or has no user
  const hasUser = data?.user && data !== null;
  
  return {
    user: data?.user,
    isAuthenticated: !!hasUser && !actualError,
    isLoading,
    error: actualError,
  };
};

// Signup mutation
export const useSignup = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: SignupRequest) => apiClient.signup(data),
    onSuccess: (response) => {
      // Set user data in cache
      queryClient.setQueryData(authKeys.currentUser(), { user: response.user });
      
      // Show success message
      toaster.create({
        title: "Account created successfully!",
        description: `Welcome, ${response.user.username}!`,
        type: "success",
      });
    },
    onError: (error) => {
      toaster.create({
        title: "Signup failed",
        description: error.message,
        type: "error",
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
      queryClient.setQueryData(authKeys.currentUser(), { user: response.user });

      // Show success message
      toaster.create({
        title: "Signed in successfully!",
        description: `Welcome back, ${response.user.username}!`,
        type: "success",
      });
    },
    onError: (error) => {
      toaster.create({
        title: "Sign in failed",
        description: error.message,
        type: "error",
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
      // Explicitly set the current user data to null
      queryClient.setQueryData(authKeys.currentUser(), null);
      
      // Clear all auth data
      queryClient.removeQueries({ queryKey: authKeys.currentUser() });
      queryClient.removeQueries({ queryKey: authKeys.users() });

      // Show success message
      toaster.create({
        title: "Signed out successfully",
        description: "See you later!",
        type: "info",
      });
    },
    onError: (error) => {
      toaster.create({
        title: "Sign out failed",
        description: error.message,
        type: "error",
      });
    },
  });
};
