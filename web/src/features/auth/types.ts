export type LoginResp = {
  expires_in: number;
};

export type RegisterConflictResp = {
  is_username_conflict: boolean;
  is_email_conflict: boolean;
};

export type CurrentUser = {
  id: number;
  username: string;
  email: string;
  role: "user" | "admin";
};
