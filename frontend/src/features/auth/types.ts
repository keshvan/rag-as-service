export type LoginPayload = {
  email: string;
  password: string;
};

export type RegisterPayload = LoginPayload & {
  organization_name: string;
  organization_url: string;
};

export type TokenResponse = {
  guid?: string;
  access_token: string;
  refresh_token: string;
};

export type RegisterResponse = {
  message: string;
};

export type UserResponse = {
  guid: string;
};

export type SessionUser = {
  guid: string;
};
