"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.SKYClient = void 0;
const axios_1 = __importDefault(require("axios"));
class SKYClient {
    constructor(baseURL = 'https://aws-kyc-verification.onrender.com') {
        this.api = axios_1.default.create({ baseURL });
    }
    async submitKYC(payload) {
        const formData = new FormData();
        formData.append('email', payload.email);
        formData.append('id_image', payload.idImage);
        formData.append('selfile', payload.sefile);
        try {
            const response = await this.api.post('/kyc', formData, {
                headers: { 'Content-Type': 'multipart/form-data' },
            });
            return response.data;
        }
        catch (error) {
            return {
                success: false,
                message: error?.response?.data?.message || 'KYC submission failed',
                data: error?.response?.data,
            };
        }
    }
}
exports.SKYClient = SKYClient;
//# sourceMappingURL=ASwan.js.map