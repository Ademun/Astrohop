interface ApiErrorBody {
    error: string;
}

interface UnauthorizedErrorBody {
    message: string;
}

export type {ApiErrorBody, UnauthorizedErrorBody}