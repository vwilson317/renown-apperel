import axios from 'axios';

export enum Api {
    Finding,
    Shopping,
}

const findApiUrl = 'api/finding';
const shoppingApiUrl = 'api/shopping'; // "http://open.api.ebay.com/shopping";
const shoppingApiVersion = 1099; // 1085 doesn't work now wtf
const eBayAppId = 'VincentW-renownap-PRD-0b31f104d-07a63429';

export const getListingItemsByStore = async (pageNum: number): Promise<any> => {
    const result = await get(Api.Finding, 'pageNum=' + pageNum);
    return result;
};

export const getItemDetails = async (itemIds: string): Promise<any> => {
    const result = await get(Api.Shopping, 'callname=GetMultipleItems&ItemId=' + itemIds);
    return result;
};

export const get = async (apiType: Api, singleParam: string): Promise<any> => {
    // todo add ebay apis to dev server

    let uri = '';
    if (apiType === Api.Finding) {
        return axios.get(findApiUrl + '?' + singleParam);
    } else if (apiType === Api.Shopping) {
        const str = 'version=' + shoppingApiVersion + '&appid=' + eBayAppId + '&responseencoding=JSON';
        uri = shoppingApiUrl + '?' + str + '&' + singleParam;

        return axios.get(uri);
    }
};

