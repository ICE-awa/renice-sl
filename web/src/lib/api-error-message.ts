const API_ERROR_MESSAGES: Record<number, string> = {
  400001: "请求参数不正确",
  400002: "邮箱验证码错误",
  400003: "账号不存在",
  400004: "密码错误",
  401000: "登录已失效，请重新登录",
  401001: "登录已失效，请重新登录",
  403000: "没有权限执行此操作",
  404000: "资源不存在",
  404001: "用户不存在",
  409001: "用户名或邮箱已被占用",
  500000: "服务器暂时不可用",
  503000: "服务暂时不可用，请稍后再试",
};

export function getApiErrorMessage(code: number, fallback: string) {
  return API_ERROR_MESSAGES[code] ?? fallback;
}
