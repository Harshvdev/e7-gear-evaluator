# Official Equipment Score (ES)

$$ES = \left\lfloor \text{Main Stat Contribution} + \sum (\text{Substat Value} \times \text{Official Weight}) \right\rfloor$$

## Official Weights & Values

### Main Stat
- **% Stats, CC, CD:** $\text{Value} \div 2.5$
- **Speed:** $\text{Value} \times 2.0$
- **Flat Stat (Lvl 85 maxed):** Fixed $24.2$ points

### Substats
- **Attack% / Health% / Defense% / Effectiveness / Effect Resistance:** $1.0$ per $1\%$
- **Speed:** $2.0$ per $1$ point
- **Critical Hit Chance:** $1.6$ per $1\%$
- **Critical Hit Damage:** $1.142857$ ($\frac{8}{7}$) per $1\%$
- **Flat Attack:** $\text{Value} \times \frac{3.7373}{39.4717} \approx 0.09468$
- **Flat Health:** $\text{Value} \times \frac{3.1505}{179.4031} \approx 0.01756$
- **Flat Defense:** $\text{Value} \times \frac{4.9853}{31.0051} \approx 0.16079$

---

## ES Example (Level 85 Ring, +9)

### Gear Details
- **Main Stat:** $34\%$ Defense
- **Substats:** $82$ Flat Attack, $7\%$ Health, $20\%$ Critical Hit Damage, $8\%$ Effectiveness

### Calculation
- **Main Stat Score:** $34 \div 2.5 = 13.60$
- **Substats Score:** 
  $$(82 \times 0.09468) + (7 \times 1.0) + \left(20 \times \frac{8}{7}\right) + (8 \times 1.0) = 7.76 + 7.00 + 22.86 + 8.00 = 45.62$$
- **Total ES:** 
  $$\text{Total ES} = \lfloor 13.60 + 45.62 \rfloor = \lfloor 59.22 \rfloor = 59$$

---

## Official In-Game Equipment Score (ES) Example

The official game engine includes the main stat (which scales to $44\%$ at level 85, +12 enhancement) and uses precise decimal scaling for flat stats, flooring the final result.

### Main Stat Contribution
- **Attack (44%):** $44 \div 2.5 = 17.60$

### Substat Contribution
- **Critical Hit Damage (7%):** $7 \times \frac{8}{7} = 8.00$
- **Effectiveness (22%):** $22 \times 1.0 = 22.00$
- **Speed (12):** $12 \times 2.0 = 24.00$
- **Flat Health (179):** $179 \times \frac{3.1505}{179.4031} \approx 3.15$

### Total Score
$$\text{Total ES} = \lfloor 17.60 + 8.00 + 22.00 + 24.00 + 3.15 \rfloor = \lfloor 74.75 \rfloor = \mathbf{74}$$
