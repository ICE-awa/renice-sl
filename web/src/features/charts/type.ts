export type StatRequest = {
  range: number;
  bucket: "hour" | "day";
};

export type StatResponse = {
  time: string;
  count: number;
};
