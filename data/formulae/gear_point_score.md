# Gear Point Score (GSP)

Gear Point Score (GSP) evaluates the stat progression of a character based on the bonus stats gained beyond their base values, multiplied by their respective stat weights.

$$\text{Total GSP} = \sum \text{GSP}_{\text{stat}}$$

## Calculating GSP for Each Stat

- **Attack GSP:**
  $$\text{Bonus } \% = \left( \frac{\text{Current Atk} - \text{Base Atk}}{\text{Base Atk}} \right) \times 100 = \left( \frac{4440 - 1222}{1222} \right) \times 100 = 263.3\%$$
  $$\text{GSP}_{\text{Atk}} = 263.3 \times 1.0 = \mathbf{263.3}$$

- **Crit Damage GSP:**
  $$\text{Bonus Value} = \text{Current CD} - \text{Base CD} = 347\% - 150\% = 197\%$$
  $$\text{GSP}_{\text{CD}} = 197 \times 1.14 = \mathbf{224.6}$$

- **Health GSP:**
  $$\text{Bonus } \% = \left( \frac{\text{Current HP} - \text{Base HP}}{\text{Base HP}} \right) \times 100 = \left( \frac{13113 - 5784}{5784} \right) \times 100 = 126.7\%$$
  $$\text{GSP}_{\text{HP}} = 126.7 \times 1.0 = \mathbf{126.7}$$

- **Speed GSP:**
  $$\text{Bonus Value} = \text{Current Speed} - \text{Base Speed} = 183 - 115 = 68 \text{ Speed}$$
  $$\text{GSP}_{\text{Spd}} = 68 \times 2.0 = \mathbf{136.0}$$

- **Defense GSP:**
  $$\text{Bonus } \% = \left( \frac{\text{Current Def} - \text{Base Def}}{\text{Base Def}} \right) \times 100 = \left( \frac{1062 - 652}{652} \right) \times 100 = 62.8\%$$
  $$\text{GSP}_{\text{Def}} = 62.8 \times 1.0 = \mathbf{62.8}$$

- **Crit Chance GSP:**
  $$\text{Bonus Value} = \text{Current CC} - \text{Base CC} = 39\% - 15\% = 24\%$$
  $$\text{GSP}_{\text{CC}} = 24 \times 1.6 = \mathbf{38.4}$$
