export type UserID = number;

export type Users = UserID[];

export interface User {
  ID: UserID;
  Name: string;
  Email: string;
}

export interface Session {
  UserID: UserID;
  Token: string;
}

export interface SignupBody {
  User: User;
  Password: string;
}

export interface SignupResponse {
  User: User;
}

export interface SignInBody {
  UserID: UserID;
  Password: string;
}

export interface SignInResponse {
  Session: Session;
}

export interface GetUserResponse {
  User: User;
}

export interface GetUsersBody {
  UserIDs: UserID[];
  test: string;
}

export interface GetUsersResponse {
  Users: Record<UserID, User>;
}

export interface SignOutBody {
  Session: Session;
}

export interface SignOutResponse {
}

// API client for GonzoServer
export const Signup = async (body: SignupBody): Promise<SignupResponse> => {
  const response = await fetch(`/user/new`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json() as Promise<SignupResponse>;
};

export const SignIn = async (body: SignInBody): Promise<SignInResponse> => {
  const response = await fetch(`/session/new`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json() as Promise<SignInResponse>;
};

export const SignOut = async (body: SignOutBody): Promise<SignOutResponse> => {
  const response = await fetch(`/session`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json() as Promise<SignOutResponse>;
};

export const GetUser = async (UserID: string): Promise<GetUserResponse> => {
  const response = await fetch(`/user/${UserID}`, {
    method: 'GET',
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json() as Promise<GetUserResponse>;
};

export const GetUsers = async (body: GetUsersBody): Promise<GetUsersResponse> => {
  const response = await fetch(`/users`, {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json() as Promise<GetUsersResponse>;
};

