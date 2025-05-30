import axios, {AxiosInstance} from 'axios';

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

export class SKYClient {
    private api: AxiosInstance; 
    constructor(baseURL: string = 'https://aws-kyc-verification.onrender.com'){
        this.api = axios.create({baseURL});
    }
    
    async submitKYC(payload: KYCRequest): Promise<KYCResponse>{
        const formData = new FormData();
        formData.append('email', payload.email);
        formData.append('id_image', payload.idImage);
        formData.append('selfile', payload.sefile);
        try{
            const response = await this.api.post<KYCResponse>('/kyc', formData, {
                headers: { 'Content-Type': 'multipart/form-data' },
              });
              return response.data;
        }catch(error: any){
            return {
                success: false,
                message: error?.response?.data?.message || 'KYC submission failed',
                data: error?.response?.data,
            }
        }
    }
}