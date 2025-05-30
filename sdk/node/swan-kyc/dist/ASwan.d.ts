export interface KYCRequest {
    email: string;
    idImage: File | Blob;
    sefile: File | Blob;
}
export interface KYCResponse {
    success: boolean;
    message?: string;
    data?: any;
}
export declare class SKYClient {
    private api;
    constructor(baseURL?: string);
    submitKYC(payload: KYCRequest): Promise<KYCResponse>;
}
