export type LinkItem = {
  id: number;
  originalUrl: string;
  shortUrl: string;
  viewCount: number;
  createdAt: string;
  updatedAt: string;
  expiresAt: string;
};

export type SidebarItems = {
  id: number;
  name: string;
  next: string;
};
