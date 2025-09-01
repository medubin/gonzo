import { GonzoServerClient } from "./client";

// Create client instance
export const apiClient = new GonzoServerClient("http://localhost:8080");

// Export types for easy importing
export * from "./types";
export { GonzoServerClient } from "./client";