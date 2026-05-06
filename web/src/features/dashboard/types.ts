export type LinkItem = {
  id: number;
  original_url: string;
  code: string;
  view_count: number;
  status: "active" | "inactive";
  created_at: string;
  updated_at: string;
  expires_at: string;
};

export type SidebarItems = {
  id: number;
  name: string;
  next: string;
};

// Api Request
export type CreateLinkFormValues = {
  original_url: string;
  expires_at?: Date;
};

export type CreateLinkInput = {
  original_url: string;
  expires_at?: string;
};

export type GetLinksInput = {
  original_url?: string;
  code?: string;
  status?: string;
  expires_begin?: string;
  expires_end?: string;
  page_num: number;
  page_size: number;
};

export type UpdateLinkFormValues = {
  id: number;
  original_url?: string;
  expires_at?: Date;
  enabled: boolean;
};

export type UpdateLinkInput = {
  id: number;
  original_url?: string;
  expires_at?: string;
  status: "active" | "inactive";
};

// Api Response
export type GetStatsResponse = {
  link_count: number;
  view_count: number;
};
