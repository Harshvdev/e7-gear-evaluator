# Community Gear Score (GS)

$$GS = \sum (\text{Substat Value} \times \text{Community Weight})$$

> [!NOTE]
> Main stats are completely ignored in community GS calculation.

## Community Weights (Standard Fribbels Optimizer Defaults)

- **Attack% / Health% / Defense% / Effectiveness / Effect Resistance:** $1.0$ per $1\%$
- **Speed:** $2.0$ per $1$ point
- **Critical Hit Chance:** $1.6$ per $1\%$
- **Critical Hit Damage:** $1.14$ per $1\%$ *(Note: standard legacy variants sometimes use 1.1)*
- **Flat Attack:** $\text{Value} \div 39$
- **Flat Health:** $\text{Value} \div 179$
- **Flat Defense:** $\text{Value} \div 31$

---

## GS Example 1 (Flat/Percent Stats Mix)

### Substats
- $82$ Flat Attack
- $7\%$ Health
- $20\%$ Critical Hit Damage
- $8\%$ Effectiveness

### Calculation
- **Flat Attack:** $82 \div 39 \approx 2.10$
- **Health %:** $7 \times 1.0 = 7.00$
- **Critical Hit Damage:** $20 \times 1.14 = 22.80$
- **Effectiveness:** $8 \times 1.0 = 8.00$
- **Total GS:** 
  $$\text{Total GS} = 2.10 + 7.00 + 22.80 + 8.00 = 39.90$$

---

## GS Example 2

The community score ignores the main stat entirely and calculates only the raw value of the substats.

### Substats
- $7\%$ Critical Hit Damage
- $22\%$ Effectiveness
- $12$ Speed
- $179$ Flat Health

### Calculation
- **Critical Hit Damage:** $7 \times 1.14 = 7.98$
- **Effectiveness:** $22 \times 1.0 = 22.00$
- **Speed:** $12 \times 2.0 = 24.00$
- **Flat Health:** $179 \div 179 = 1.00$
- **Total GS:** 
  $$\text{Total GS} = 7.98 + 22.00 + 24.00 + 1.00 = 54.98$$