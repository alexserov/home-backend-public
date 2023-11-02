type DiffObj<T> = {
    changed: boolean;
    value: Diff<T>;
}

type DiffItem<T> = T extends object ? DiffObj<T> : {
    changed: boolean;
    oldValue: T;
    newValue: T;
}
export type Diff<T> = {
    [key in keyof T]: DiffItem<T[key]>;
};

export function buildDiff<T extends Record<string|number, any>>(oldValue:T, newValue:T): Diff<T> {
    const result: Diff<T> = {} as unknown as Diff<T>;
    for(const key of Object.keys(oldValue)) {
        const oldPropValue = oldValue[key];
        const newPropValue = newValue[key];
        if(typeof oldPropValue !== 'object') {
            if(oldPropValue!==newPropValue){
                (result as any)[key] = {
                    changed: true,
                    oldValue: oldPropValue,
                    newValue: newPropValue,
                }
            } else {
                (result as any)[key] = {
                    changed: false,
                    oldValue: oldPropValue,
                    newValue: newPropValue,
                }
            }
        } else {
            const propDiff = buildDiff(oldPropValue, newPropValue);
            (result as any)[key] = {
                changed: Object.values(propDiff).some(x=>x.changed),
                value: propDiff
            };
        }
    }
    return result;
}