export type GetDLQMessagesParams = {
  page_num: number;
  page_size: number;
};

export type DLQMessageItem = {
  id: number;
  source_stream: string;
  source_consumer: string;
  stream_seq: number;
  subject: string;
  payload: unknown;
  reason: string;
  status: string;
  failed_at: string;
  resolved_at: string | null;
};

export type GetDLQMessagesResponse = {
  total: number;
  items: DLQMessageItem[];
  page_num: number;
  page_size: number;
};
